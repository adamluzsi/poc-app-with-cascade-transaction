package httpapi

import (
	"github.com/adamluzsi/frameless/pkg/txs"
	app "github.com/adamluzsi/poc-app-with-cascade-transaction"
	"net/http"
)

func NewHandler(uc app.UseCase) http.Handler {
	m := http.NewServeMux()
	h := Handler{UseCase: uc}
	mw := TxMiddleware{Next: h}
	m.Handle("/", mw)
	return m
}

type Handler struct {
	app.UseCase
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ent := app.Entity{ // map entity from DTO
		V: 42,
	}
	if err := h.UseCase.Do(r.Context(), ent); err != nil {
		http.Error(w, "boom", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type TxMiddleware struct {
	Next http.Handler
}

func (mw TxMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, _ := txs.Begin(r.Context())
	rwp := &ResponseWriterProxy{ResponseWriter: w}
	mw.Next.ServeHTTP(rwp, r.WithContext(ctx))
	if 200 <= rwp.Code && rwp.Code < 300 {
		_ = txs.Commit(ctx)
	} else {
		_ = txs.Rollback(ctx)
	}
}

type ResponseWriterProxy struct {
	http.ResponseWriter
	Code int
}

func (rwp *ResponseWriterProxy) WriteHeader(statusCode int) {
	rwp.Code = statusCode
	rwp.ResponseWriter.WriteHeader(statusCode)
}
