package api

import (
	"errors"
	"fmt"
	"net/http"

	util "github.com/qri-io/apiutil"
	"github.com/qri-io/qri/config"
	"github.com/qri-io/qri/fsi"
	"github.com/qri-io/qri/lib"
	"github.com/qri-io/qri/repo"
	"github.com/qri-io/qri/repo/profile"
	reporef "github.com/qri-io/qri/repo/ref"
)

// RootHandler bundles handlers that may need to be called
// by "/"
// TODO - This will be removed when we add a proper router
type RootHandler struct {
	dsh *DatasetHandlers
	ph  *PeerHandlers
}

// NewRootHandler creates a new RootHandler
func NewRootHandler(dsh *DatasetHandlers, ph *PeerHandlers) *RootHandler {
	return &RootHandler{dsh, ph}
}

// Handler is the only Handler func for the RootHandler struct
func (mh *RootHandler) Handler(w http.ResponseWriter, r *http.Request) {
	ref := DatasetRefFromCtx(r.Context())
	if ref.IsEmpty() {
		HealthCheckHandler(w, r)
		return
	}

	if ref.IsPeerRef() {
		p := &lib.PeerInfoParams{
			Peername: ref.Peername,
		}
		res := &config.ProfilePod{}
		err := mh.ph.Info(p, res)
		if err != nil {
			util.WriteErrResponse(w, http.StatusInternalServerError, err)
			return
		}
		if res.ID == "" {
			util.WriteErrResponse(w, http.StatusNotFound, errors.New("cannot find peer"))
			return
		}
		util.WriteResponse(w, res)
		return
	}

	format := r.FormValue("format")
	selector := r.FormValue("component")
	p := lib.GetParams{
		Refstr:   ref.String(),
		Format:   format,
		Selector: selector,
	}
	res := lib.GetResult{}
	err := mh.dsh.Get(&p, &res)
	if err != nil {
		if err == repo.ErrNotFound {
			util.NotFoundHandler(w, r)
			return
		}
		if err == fsi.ErrNoLink {
			util.WriteErrResponse(w, http.StatusUnprocessableEntity, err)
			return
		}
		util.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	if res.Dataset == nil || res.Dataset.IsEmpty() {
		util.WriteErrResponse(w, http.StatusNotFound, errors.New("cannot find peer dataset"))
		return
	}

	if err := inlineScriptsToBytes(res.Dataset); err != nil {
		util.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	if format == "zip" {
		zipFilename := fmt.Sprintf("%s.zip", ref.Name)
		w.Header().Set("Content-Type", extensionToMimeType(".zip"))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment: filename=%s", zipFilename))
		w.Write(res.Bytes)
		return
	}

	// TODO (b5) - why is this necessary?
	ref = reporef.DatasetRef{
		Peername:  res.Dataset.Peername,
		ProfileID: profile.IDB58DecodeOrEmpty(res.Dataset.ProfileID),
		Name:      res.Dataset.Name,
		Path:      res.Dataset.Path,
		FSIPath:   res.FSIPath,
		Published: res.Published,
		Dataset:   res.Dataset,
	}

	util.WriteResponse(w, ref)
	return
}
