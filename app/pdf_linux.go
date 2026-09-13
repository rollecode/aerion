//go:build linux

package app

/*
#cgo linux pkg-config: gtk+-3.0 gtk+-unix-print-3.0
#cgo !webkit2_41 pkg-config: webkit2gtk-4.0
#cgo webkit2_41 pkg-config: webkit2gtk-4.1

#include <gtk/gtk.h>
#include <gtk/gtkunixprint.h>
#include <webkit2/webkit2.h>
#include <stdlib.h>
#include <string.h>

// Renders HTML to PDF through an offscreen WebKit view so the output matches
// what the print dialog would produce. All GTK work happens on the main loop.
typedef struct {
	char      *html;
	char      *uri;
	GMutex     lock;
	GCond      cond;
	gboolean   done;
	char      *err;
	WebKitWebView *view;
	gboolean   cleaned;
	gboolean   terminated;
	int        refs;
} pdfJob;

static void pdf_job_unref(pdfJob *j) {
	g_mutex_lock(&j->lock);
	j->refs--;
	gboolean dead = j->refs <= 0;
	g_mutex_unlock(&j->lock);
	if (!dead) {
		return;
	}
	g_free(j->html);
	g_free(j->uri);
	g_free(j->err);
	g_mutex_clear(&j->lock);
	g_cond_clear(&j->cond);
	g_free(j);
}

// Wakes the waiting caller. The first settle wins, so a failure reported
// before "finished" keeps its error message.
static void pdf_job_settle(pdfJob *j, const char *err) {
	g_mutex_lock(&j->lock);
	if (!j->done) {
		j->done = TRUE;
		if (err != NULL) {
			j->err = g_strdup(err);
		}
		g_cond_signal(&j->cond);
	}
	g_mutex_unlock(&j->lock);
}

// Must only run on the GTK main loop, which also makes the guard race-free.
static void pdf_job_cleanup(pdfJob *j) {
	if (j->cleaned) {
		return;
	}
	j->cleaned = TRUE;
	if (j->view != NULL) {
		g_object_unref(j->view);
		j->view = NULL;
	}
}

// Ends the GTK side exactly once. Main-loop only, so the flag needs no lock.
static void pdf_job_terminate(pdfJob *j) {
	if (j->terminated) {
		return;
	}
	j->terminated = TRUE;
	pdf_job_cleanup(j);
	pdf_job_unref(j);
}

static void pdf_on_finished(WebKitPrintOperation *op, gpointer data) {
	pdfJob *j = (pdfJob *)data;
	pdf_job_settle(j, NULL);
	g_object_unref(op);
	pdf_job_terminate(j);
}

// GTK presents the print-to-file printer under a translated display name, so
// it is located by backend type instead. NULL when that backend is disabled.
static gboolean pdf_match_file_printer(GtkPrinter *printer, gpointer data) {
	GtkPrintBackend *backend = gtk_printer_get_backend(printer);
	if (backend == NULL || g_strcmp0(G_OBJECT_TYPE_NAME(backend), "GtkPrintBackendFile") != 0) {
		return FALSE;
	}
	*(GtkPrinter **)data = g_object_ref(printer);
	return TRUE;
}

static char *pdf_file_printer_name(void) {
	GtkPrinter *found = NULL;
	gtk_enumerate_printers(pdf_match_file_printer, &found, NULL, TRUE);
	if (found == NULL) {
		return NULL;
	}
	char *name = g_strdup(gtk_printer_get_name(found));
	g_object_unref(found);
	return name;
}

static void pdf_on_failed(WebKitPrintOperation *op, GError *err, gpointer data) {
	pdfJob *j = (pdfJob *)data;
	pdf_job_settle(j, (err != NULL && err->message != NULL) ? err->message : "print operation failed");
}

static void pdf_on_load(WebKitWebView *view, WebKitLoadEvent ev, gpointer data) {
	if (ev != WEBKIT_LOAD_FINISHED) {
		return;
	}
	pdfJob *j = (pdfJob *)data;

	char *printer = pdf_file_printer_name();
	if (printer == NULL) {
		pdf_job_settle(j, "GTK print-to-file backend is unavailable");
		pdf_job_terminate(j);
		return;
	}

	WebKitPrintOperation *op = webkit_print_operation_new(view);
	GtkPrintSettings *settings = gtk_print_settings_new();
	gtk_print_settings_set_printer(settings, printer);
	g_free(printer);
	gtk_print_settings_set(settings, GTK_PRINT_SETTINGS_OUTPUT_URI, j->uri);
	gtk_print_settings_set(settings, GTK_PRINT_SETTINGS_OUTPUT_FILE_FORMAT, "pdf");
	webkit_print_operation_set_print_settings(op, settings);
	g_object_unref(settings);

	g_signal_connect(op, "finished", G_CALLBACK(pdf_on_finished), j);
	g_signal_connect(op, "failed", G_CALLBACK(pdf_on_failed), j);
	webkit_print_operation_print(op);
}

// The view is deliberately left unparented. Hosting it in a GtkOffscreenWindow
// aborts under Wayland, which cannot hand that window a GL context.
static gboolean pdf_start(gpointer data) {
	pdfJob *j = (pdfJob *)data;
	j->view = WEBKIT_WEB_VIEW(webkit_web_view_new());
	g_object_ref_sink(j->view);
	g_signal_connect(j->view, "load-changed", G_CALLBACK(pdf_on_load), j);
	webkit_web_view_load_html(j->view, j->html, NULL);
	return G_SOURCE_REMOVE;
}

// Blocks the calling thread until the PDF is written. Returns NULL on success,
// otherwise a newly allocated error string the caller must free.
static char *pdf_render(const char *html, const char *uri, int timeoutSeconds) {
	pdfJob *j = g_new0(pdfJob, 1);
	g_mutex_init(&j->lock);
	g_cond_init(&j->cond);
	j->html = g_strdup(html);
	j->uri = g_strdup(uri);
	j->refs = 2;

	g_idle_add(pdf_start, j);

	gint64 deadline = g_get_monotonic_time() + (gint64)timeoutSeconds * G_TIME_SPAN_SECOND;
	g_mutex_lock(&j->lock);
	while (!j->done) {
		if (!g_cond_wait_until(&j->cond, &j->lock, deadline)) {
			j->done = TRUE;
			j->err = g_strdup("timed out rendering PDF");
			break;
		}
	}
	char *err = (j->err != NULL) ? strdup(j->err) : NULL;
	g_mutex_unlock(&j->lock);

	pdf_job_unref(j);
	return err;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// pdfRenderTimeoutSeconds bounds a single render so a stuck page cannot hang
// the caller forever.
const pdfRenderTimeoutSeconds = 60

func renderHTMLToPDF(html, path string) error {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))

	cURI := C.CString("file://" + path)
	defer C.free(unsafe.Pointer(cURI))

	cErr := C.pdf_render(cHTML, cURI, C.int(pdfRenderTimeoutSeconds))
	if cErr == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cErr))
	return fmt.Errorf("%s", C.GoString(cErr))
}
