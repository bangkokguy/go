# ThermoMan

## OpenAPI 3.0 Generated Go with Chi framework support

This project contains generated code from an OpenAPI 3 definition. The project includes the Go Chi library and is ready to run stand alone as a service.

### How to build

To build this project, a Go installation is required beforehand. You may then execute:

> go build .

or

> make

### How to run

To run the application:

> go run .

or

> make run

### Generated code

Typically as an API is iterated on, the code is regenerated to stay in sync with the API definition. However, once a developer starts adding their own code to the stub implementation class(es), they wont want to overwrite those file(s). To avoid this, when unzipping the generated archive file, avoid overwriting the stub implementation files.

The Go generated project has a few parts to it. The api/ directory contains one file per unique path and a delegate file that is a structure with delegate references per unique path. Each generated file represents all of the operations a unique path contains as defined in the API definition. These functions are complete implementations of handler functions matching the standard net/http.Handle function parameter signature. These functions are all referenced in the router generated source file to map their paths to the functions. These are the functions that are called when requests are made to this service. In turn, these functions will pull out any API defined parameters, reducing the coding you will need to do in your implementation for each function. Using a delegate pattern, each of the generated functions will call your implementation function of the same signature, but include any parameters as a 'param' option.

The router.go source found in router/router.go... as mentioned above contains all the mappings of every path and operation to the correct function in the api/* implemented functions. Any time the API is modified and code regenerated, be sure to overwrite this file as it will be updated to match changes/additions/removal of paths and operations.

In the types/ directory you will find generated source code that matches any defined components as well as unique path specific parameter source that contain one or more structures.. one for each path that has defined at least one type of parameter for the operation (header, path param, query param or cookie).

The main.go source is auto-generated as well and should not need to be modified. It contains the necessary code to start the API service and sets up all the delegate functions per unique path defined in your API. You may find it necessary to modify the port it starts on.

### Implementation... your code goes here

For each unique path, you will find a generated go source file in the impl/ directory. In each of these files you will find stubbed out functions which will have the request and response attributes passed to them. In cases where path, header, query or cookie parameters are defined for the path that a function represents, it will also have a parameter called 'params' that is a generated structure which will contain fields that match the name and type of each defined parameter for the operation the generated function matches.

These generated stub functions are provided as a means to get you started on implementing your API. Once you add your own code, you won't want these functions overwritten. When changes are made to the API definition and code regenerated, you will not want to overwrite these stub functions when you extract the generated archive so as not to have to copy/paste the code you add back in to newly created empty stub functions. As a benefit, when you attempt to build your project, if your current implementation does not match any changes that were made to the API definition, your project may not build. Your IDE may also flag any issues if it is capable and set up in that manner. The benefit is it alerts you that changes were made and you will need to update your implementation to stay in sync with the API changes.

The request object is provided should you need to access any aspect of the request. As well, should you need to access the request body, you may do so with this request object.

### Test a request

Once the server is running, noting the port you started it at, making a request using curl if installed:

> curl --header "Accept: application/json" http://localhost:<PORT>/path

where <PORT> matches the port you started the service on, and /path matches a path defined in your OpenAPI definition.