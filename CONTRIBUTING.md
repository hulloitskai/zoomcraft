# Contributing to Zoomcraft

Wanna help me make this thing better? Dope!

Please make a pull request and tag me to review it. Feel free to email me
at <hello@stevenxie.me> if you're not sure where to get started.

Some general guidelines to keep in mind:

- Let's keep things clean. Run `yarn lint` on `client`, and
  `go fmt ./... && go vet ./...` on `backend` before you push.

- **Node**
  - Try to keep development dependencies in `package.json/devDependencies`,
    and only install into `package.json/dependencies` if the library is
    required at run-time for end users.
- **Docker**
  - Please keep the final build slim! Use builder images and copy build
    products from them into the final image.
