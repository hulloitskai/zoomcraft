package graphqlutil

import (
	"html/template"
	"net/http"

	"github.com/cockroachdb/errors"
)

const giqlVersion = "0.11.11"

// ServeGraphiQL creates an http.HandlerFunc that serves the GraphiQL GUI.
func ServeGraphiQL(endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			t, err := template.New("GraphiQL").Parse(giqlTemplate)
			if err != nil {
				return errors.Wrap(err, "parsing template")
			}

			// Parse query params.
			param := r.URL.Query()
			query := param.Get("query")
			if query == "" {
				query = defaultQuery
			}

			d := giqlData{
				Version:         giqlVersion,
				Endpoint:        endpoint,
				QueryString:     query,
				VariablesString: param.Get("variables"),
				OperationName:   param.Get("operationName"),
			}
			if err = t.Execute(w, d); err != nil {
				return errors.Wrap(err, "executing template")
			}
			return nil
		}(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// giqlData is the page data structure of the rendered GraphiQL page
type giqlData struct {
	Version         string
	Endpoint        string
	QueryString     string
	VariablesString string
	OperationName   string
}

const giqlTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>GraphiQL</title>
  <meta name="robots" content="noindex" />
  <meta name="referrer" content="origin">
  <style>
    body {
      height: 100%;
      margin: 0;
      overflow: hidden;
      width: 100%;
    }
    #graphiql {
      height: 100vh;
    }
  </style>
  <link href="https://cdn.jsdelivr.net/npm/graphiql@{{ .Version }}/graphiql.css" rel="stylesheet" />
  <script src="https://cdn.jsdelivr.net/es6-promise/4.0.5/es6-promise.auto.min.js"></script>
  <script src="https://cdn.jsdelivr.net/fetch/0.9.0/fetch.min.js"></script>
  <script src="https://cdn.jsdelivr.net/react/15.4.2/react.min.js"></script>
  <script src="https://cdn.jsdelivr.net/react/15.4.2/react-dom.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/graphiql@{{ .Version }}/graphiql.min.js"></script>
</head>
<body>
  <div id="graphiql">Loading...</div>
  <script>
    // Collect the URL parameters.
    var parameters = {};
    location.search.substr(1).split('&').forEach(function (entry) {
      var eq = entry.indexOf('=');
      if (eq >= 0) {
        parameters[decodeURIComponent(entry.slice(0, eq))] =
          decodeURIComponent(entry.slice(eq + 1));
      }
    });

    // When the query and variables string is edited, update the URL bar so
    // that it can be easily shared.
    function onEditQuery(newQuery) {
      parameters.query = newQuery;
      updateURL();
    }
    function onEditVariables(newVariables) {
      parameters.variables = newVariables;
      updateURL();
    }
    function onEditOperationName(newOperationName) {
      parameters.operationName = newOperationName;
      updateURL();
    }
    function updateURL() {
      history.replaceState(null, null, locationQuery(parameters));
    }

    // Produce a Location query string from a parameter object.
    function locationQuery(params) {
      return '?' + Object.keys(params).filter(function (key) {
        return Boolean(params[key]);
      }).map(function (key) {
        return encodeURIComponent(key) + '=' +
          encodeURIComponent(params[key]);
      }).join('&');
    }

    // Derive a fetch URL from the current URL, sans the GraphQL parameters.
    var graphqlParamNames = {
      query: true,
      variables: true,
      operationName: true
    };
    var otherParams = {};
    for (var k in parameters) {
      if (parameters.hasOwnProperty(k) && graphqlParamNames[k] !== true) {
        otherParams[k] = parameters[k];
      }
    }

    // Define a HTTP GraphQL fetcher.
    var fetchURL = '{{ .Endpoint }}' + locationQuery(otherParams);
    function graphQLFetcher(graphQLParams) {
      return fetch(fetchURL, {
        method: 'post',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(graphQLParams),
        credentials: 'include',
      }).then(function (response) {
        return response.text();
      }).then(function (responseBody) {
        try {
          return JSON.parse(responseBody);
        } catch (error) {
          return responseBody;
        }
      });
    }

    // Render <GraphiQL /> into the body.
    ReactDOM.render(
      React.createElement(GraphiQL, {
        fetcher: graphQLFetcher,
        onEditQuery: onEditQuery,
        onEditVariables: onEditVariables,
        onEditOperationName: onEditOperationName,
        query: {{ .QueryString }},
        variables: {{ .VariablesString }},
        operationName: {{ .OperationName }},
      }),
      document.getElementById('graphiql')
    );
  </script>
</body>
</html>
`

const defaultQuery = `# Welcome to GraphiQL
#
# GraphiQL is an in-browser tool for writing, validating, and
# testing GraphQL queries.
#
# Type queries into this side of the screen, and you will see intelligent
# typeaheads aware of the current GraphQL type schema and live syntax and
# validation errors highlighted within the text.
#
# GraphQL queries typically start with a "{" character. Lines that starts
# with a # are ignored.
#
# An example GraphQL query might look like:
#
#     {
#       field(arg: "value") {
#         subField
#       }
#     }
#
# Keyboard shortcuts:
#
#       Run Query:  Ctrl-Enter (or press the play button above)
#
#   Auto Complete:  Ctrl-Space (or just start typing)
#

`
