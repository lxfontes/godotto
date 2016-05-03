package images

import (
	"github.com/aybabtme/godotto/internal/godoutil"
	"github.com/digitalocean/godo"
	"golang.org/x/net/context"
)

// A Client can interact with the DigitalOcean Images service.
type Client interface {
	GetByID(int) (Image, error)
	GetBySlug(string) (Image, error)
	Update(int, ...UpdateOpt) (Image, error)
	Delete(int) error
	List(context.Context) (<-chan Image, <-chan error)
	ListApplication(context.Context) (<-chan Image, <-chan error)
	ListDistribution(context.Context) (<-chan Image, <-chan error)
	ListUser(context.Context) (<-chan Image, <-chan error)
}

// A Image in the DigitalOcean cloud.
type Image interface {
	Struct() *godo.Image
}

// New creates a Client.
func New(g *godo.Client) Client {
	c := &client{
		g: g,
	}
	return c
}

type client struct {
	g *godo.Client
}

func (svc *client) GetByID(id int) (Image, error) {
	d, _, err := svc.g.Images.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &image{g: svc.g, d: d}, nil
}

func (svc *client) GetBySlug(slug string) (Image, error) {
	d, _, err := svc.g.Images.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	return &image{g: svc.g, d: d}, nil
}

// UpdateOpt is an optional argument to images.Update.
type UpdateOpt func(*updateOpt)

func UseGodoImage(req *godo.ImageUpdateRequest) UpdateOpt {
	return func(opt *updateOpt) { opt.req = req }
}

type updateOpt struct {
	req *godo.ImageUpdateRequest
}

func (svc *client) defaultUpdateOpts() *updateOpt {
	return &updateOpt{
		req: &godo.ImageUpdateRequest{},
	}
}

func (svc *client) Update(id int, opts ...UpdateOpt) (Image, error) {
	opt := svc.defaultUpdateOpts()
	for _, fn := range opts {
		fn(opt)
	}
	d, _, err := svc.g.Images.Update(id, opt.req)
	if err != nil {
		return nil, err
	}
	return &image{g: svc.g, d: d}, nil
}

func (svc *client) Delete(id int) error {
	_, err := svc.g.Images.Delete(id)
	return err
}

func (svc *client) List(ctx context.Context) (<-chan Image, <-chan error) {
	return svc.listCommon(ctx, svc.g.Images.List)
}

func (svc *client) ListApplication(ctx context.Context) (<-chan Image, <-chan error) {
	return svc.listCommon(ctx, svc.g.Images.ListApplication)
}

func (svc *client) ListDistribution(ctx context.Context) (<-chan Image, <-chan error) {
	return svc.listCommon(ctx, svc.g.Images.ListDistribution)
}

func (svc *client) ListUser(ctx context.Context) (<-chan Image, <-chan error) {
	return svc.listCommon(ctx, svc.g.Images.ListUser)
}

type listfunc func(*godo.ListOptions) ([]godo.Image, *godo.Response, error)

func (svc *client) listCommon(ctx context.Context, listFn listfunc) (<-chan Image, <-chan error) {
	outc := make(chan Image, 1)
	errc := make(chan error, 1)

	go func() {
		defer close(outc)
		defer close(errc)
		err := godoutil.IterateList(ctx, func(opt *godo.ListOptions) (*godo.Response, error) {
			r, resp, err := listFn(opt)
			for _, d := range r {
				dd := d // copy ranged over variable
				select {
				case outc <- &image{g: svc.g, d: &dd}:
				case <-ctx.Done():
					return resp, err
				}
			}
			return resp, err
		})
		if err != nil {
			errc <- err
		}
	}()
	return outc, errc
}

type image struct {
	g *godo.Client
	d *godo.Image
}

func (svc *image) Struct() *godo.Image { return svc.d }
