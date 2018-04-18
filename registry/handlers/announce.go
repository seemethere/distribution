package handlers

import (
	"bytes"
	"encoding/binary"
	"net"
	"net/http"

	"github.com/docker/distribution/registry/api/v2"
	// storagedriver "github.com/docker/distribution/registry/storage/driver"
	// "github.com/docker/distribution/registry/storage/driver/middleware"
	"github.com/gorilla/handlers"
	"github.com/opencontainers/go-digest"
)

const (
	DefaultPeerPort = 6861
)

func announceDispatcher(ctx *Context, r *http.Request) http.Handler {
	dgst, err := getDigest(ctx)
	if err != nil {

		if err == errDigestNotAvailable {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx.Errors = append(ctx.Errors, v2.ErrorCodeDigestInvalid.WithDetail(err))
			})
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Errors = append(ctx.Errors, v2.ErrorCodeDigestInvalid.WithDetail(err))
		})
	}
	announceHandler := &announceHandler{
		Context: ctx,
		Digest:  dgst,
	}
	mhandler := handlers.MethodHandler{
		"GET":  http.HandlerFunc(announceHandler.GetPeers),
		"HEAD": http.HandlerFunc(announceHandler.GetPeers),
	}
	return mhandler
}

type announceHandler struct {
	*Context

	Digest digest.Digest
}

func (ah *announceHandler) GetPeers(w http.ResponseWriter, r *http.Request) {
	// TODO: use this to get our middleware that will have our list of peers
	peers := []string{}
	var body bytes.Buffer
	for _, peer := range peers {
		peerAddr := net.ParseIP(peer)
		body.Write(peerAddr)
		err := binary.Write(&body, binary.BigEndian, uint16(DefaultPeerPort))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Interval", "900")
	_, err := w.Write(body.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
