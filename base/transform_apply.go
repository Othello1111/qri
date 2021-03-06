package base

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/qri-io/dataset"
	"github.com/qri-io/ioes"
	"github.com/qri-io/qri/dsref"
	"github.com/qri-io/qri/repo"
	"github.com/qri-io/qri/startf"
)

// TODO(dustmop): Tests. Especially once the `apply` command exists.

// TransformApply applies the transform script to order to modify the changing dataset
func TransformApply(
	ctx context.Context,
	ds *dataset.Dataset,
	r repo.Repo,
	loader dsref.ParseResolveLoad,
	str ioes.IOStreams,
	scriptOut io.Writer,
	secrets map[string]string,
) error {
	pro, err := r.Profile()
	if err != nil {
		return err
	}

	var (
		target = ds
		head   *dataset.Dataset
	)

	if ds.Name != "" {
		head, err = loader(ctx, fmt.Sprintf("%s/%s", pro.Peername, ds.Name))
		if errors.Is(err, dsref.ErrRefNotFound) || errors.Is(err, dsref.ErrNoHistory) {
			// Dataset either does not exist yet, or has no history. Not an error
			head = &dataset.Dataset{}
			err = nil
		} else if err != nil {
			return err
		}
	}

	// create a check func from a record of all the parts that the datasetPod is changing,
	// the startf package will use this function to ensure the same components aren't modified
	mutateCheck := startf.MutatedComponentsFunc(target)

	opts := []func(*startf.ExecOpts){
		startf.AddQriRepo(r),
		startf.AddMutateFieldCheck(mutateCheck),
		startf.SetErrWriter(scriptOut),
		startf.SetSecrets(secrets),
		startf.AddDatasetLoader(loader),
	}

	if err = startf.ExecScript(ctx, target, head, opts...); err != nil {
		return err
	}

	str.PrintErr("✅ transform complete\n")

	return nil
}
