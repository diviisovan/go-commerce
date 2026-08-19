package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// swaggerIndexHTML is a replacement for the index page that gin-swagger
// generates, adding a requestInterceptor.
//
// Why this file exists: Swagger 2.0 (what swag v1 emits) can only describe a
// Bearer token as an opaque apiKey header, which Swagger UI sends verbatim.
// Pasting a raw JWT into the Authorize dialog therefore produces a header with
// no auth scheme, and the server correctly rejects it.
//
// The interceptor below prepends "Bearer " when it is missing, so you can paste
// just the token. Note this is purely a convenience in the UI: the request that
// actually goes over the wire still carries the standard
// "Authorization: Bearer <token>" header, and the server still requires it.
// Nothing here loosens the API contract.
//
// gin-swagger's Config has no requestInterceptor field, which is why the page is
// served by hand rather than through ginSwagger.WrapHandler alone.
const swaggerIndexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>eCommerce API - Swagger UI</title>
  <link rel="stylesheet" href="./swagger-ui.css">
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .auth-hint {
      margin: 0; padding: 10px 20px;
      background: #e8f4fd; border-bottom: 1px solid #b3d9f2;
      font-family: sans-serif; font-size: 13px; color: #1a4971;
    }
    .auth-hint code { background: #fff; padding: 1px 5px; border-radius: 3px; }
  </style>
</head>
<body>
  <p class="auth-hint">
    <strong>Authorization:</strong> paste just the <code>access_token</code> into
    <em>Authorize</em>. The <code>Bearer</code> prefix is added automatically.
  </p>
  <div id="swagger-ui"></div>
  <script src="./swagger-ui-bundle.js"></script>
  <script src="./swagger-ui-standalone-preset.js"></script>
  <script>
  window.onload = function () {
    window.ui = SwaggerUIBundle({
      url: "doc.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      plugins: [SwaggerUIBundle.plugins.DownloadUrl],
      layout: "StandaloneLayout",
      // Keeps the token across page reloads.
      persistAuthorization: true,
      requestInterceptor: function (request) {
        var headers = request.headers || {};
        // Header names are case-insensitive, so find whichever casing is present.
        var key = Object.keys(headers).filter(function (h) {
          return h.toLowerCase() === "authorization";
        })[0];

        if (key) {
          var value = String(headers[key]).trim();
          // Only add the scheme when it is genuinely absent, so a user who
          // already typed "Bearer ..." does not end up with it twice.
          if (value && !/^bearer\s/i.test(value)) {
            headers[key] = "Bearer " + value;
          }
        }
        return request;
      }
    });
  };
  </script>
</body>
</html>`

// swaggerHandler serves the Swagger UI.
//
// Registering "/swagger/index.html" alongside "/swagger/*any" would make Gin
// panic on conflicting routes, so a single wildcard handler serves the custom
// index and delegates every other path (doc.json, css, js) to gin-swagger.
func swaggerHandler(docJSONURL string) gin.HandlerFunc {
	assets := ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL(docJSONURL),
		ginSwagger.PersistAuthorization(true),
	)

	return func(c *gin.Context) {
		switch c.Param("any") {
		case "", "/", "/index.html":
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, swaggerIndexHTML)
		default:
			assets(c)
		}
	}
}
