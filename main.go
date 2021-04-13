package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/swaggest/rest"
	"github.com/swaggest/rest/chirouter"
	"github.com/swaggest/rest/jsonschema"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/swgui/v3cdn"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

func main() {
	portPtr := flag.Int("port", 3000, "webserver port")

	// Init API documentation schema.
	apiSchema := &openapi.Collector{}
	apiSchema.Reflector().SpecEns().Info.Title = "Basic Example"
	apiSchema.Reflector().SpecEns().Info.WithDescription("This app showcases a trivial REST API.")
	apiSchema.Reflector().SpecEns().Info.Version = "v1.2.3"

	// apiSchema.Reflector().SpecEns().Servers = []openapi3.Server{
	// 	openapi3.Server{
	// 		URL:         "http://www.example.com",
	// 		Description: nil,
	// 	},
	// }

	// apiSchema.Reflector().SpecEns().Tags = []openapi3.Tag{
	// 	openapi3.Tag{
	// 		Name:        "Fobbar",
	// 		Description: nil,
	// 	},
	// }

	// Setup request decoder and validator.
	validatorFactory := jsonschema.NewFactory(apiSchema, apiSchema)
	decoderFactory := request.NewDecoderFactory()
	decoderFactory.ApplyDefaults = true
	decoderFactory.SetDecoderFunc(rest.ParamInPath, chirouter.PathToURLValues)

	// Create router.
	r := chirouter.NewWrapper(chi.NewRouter())

	// Setup middlewares.
	r.Use(
		middleware.Recoverer,                          // Panic recovery.
		nethttp.OpenAPIMiddleware(apiSchema),          // Documentation collector.
		request.DecoderMiddleware(decoderFactory),     // Request decoder setup.
		request.ValidatorMiddleware(validatorFactory), // Request validator setup.
		response.EncoderMiddleware,                    // Response encoder setup.
		gzip.Middleware,                               // Response compression with support for direct gzip pass through.
	)

	greeter(r)
	badRoute(r)

	// Swagger UI endpoint at /docs.
	r.Method(http.MethodGet, "/docs/openapi.json", apiSchema)
	r.Mount("/docs", v3cdn.NewHandler(apiSchema.Reflector().Spec.Info.Title,
		"/docs/openapi.json", "/docs"))

	// Start server.
	log.Println(fmt.Sprintf("http://localhost:%v/docs", *portPtr))
	if err := http.ListenAndServe(fmt.Sprintf(":%v", *portPtr), r); err != nil {
		log.Fatal(err)
	}
}

// Declare output port type.
type helloOutput struct {
	Message string `json:"message"`
}

func greeter(r *chirouter.Wrapper) {
	// Declare input port type.
	type helloInput struct {
		Name string `path:"name"`
	}

	// Create use case interactor with references to input/output types and interaction function.
	u := usecase.NewIOI(new(helloInput), new(helloOutput), func(ctx context.Context, input, output interface{}) error {
		var (
			in  = input.(*helloInput)
			out = output.(*helloOutput)
		)

		out.Message = fmt.Sprintf("Hello %s", in.Name)

		return nil
	})

	// Describe use case interactor.
	u.SetTitle("Greeter")
	u.SetDescription("Greeter greets you.")

	u.SetExpectedErrors(status.InvalidArgument)

	// Add use case handler to router.
	r.Method(http.MethodGet, "/api/hello/{name}", nethttp.NewHandler(u))
}

func badRoute(r *chirouter.Wrapper) {
	// Declare input port type.
	type helloInput struct {
		Name string `path:"name"`
		Bad1 string `path:"bad1"`
		Bad2 string `path:"bad2"`
	}

	// Create use case interactor with references to input/output types and interaction function.
	u := usecase.NewIOI(new(helloInput), new(helloOutput), func(ctx context.Context, input, output interface{}) error {
		return status.Wrap(errors.New("bad route was called"), status.InvalidArgument)
	})

	// Describe use case interactor.
	u.SetTitle("Bad Route")
	u.SetDescription("If bad Route greets you, you found a bug.")

	u.SetExpectedErrors(status.InvalidArgument)

	r.Method(http.MethodGet, "/api/hello/{name}/{bad1}/{bad2}", nethttp.NewHandler(u))
}
