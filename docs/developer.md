# Developer Guide

## Dev Container

Install the [Dev Containers](
https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
extension for VS Code. The extension will automatically detect the
`.devcontainer` folder and prompt you to open the project in a container.

## Classic

### Install Dependencies

FastTrackML requires the following dependencies to be installed on your system:

- [Go SDK](https://go.dev/dl/)
- A working C compiler for your platform
  - macOS: `xcode-select --install`
  - Debian/Ubuntu: `sudo apt install build-essential`
  - Windows: Install [MSYS2](https://www.msys2.org)

### Build FastTrackML

FastTrackML can be built using the following command:

```bash
make build
```

## Working with the database

FastTrackML introduces database changes via migration packages, which can be found
at `pkg/database/migrations`. Each migration is given a unique package name so that
we can retain a snapshot of the `database/model.go` file that was accurate at the time
of the migration. The package names are expected to be sequential in order of application.

Make targets have been set up to make this a little less cumbersome:
```bash
make migrations-create
```
This target will create a new migration package and setup the two files you need 
(`model.go` and `migrate.go`). You will need to fill in the actual migration logic
in the `Migrate` function of the new migrate.go file -- everything else is handled by
the make target. It's assumed that `database/model.go` and `<your new migration>/model.go`
will be identical (for the time being) and include the database schema changes you want to see.

```bash
make migrations-rebuild
```
This target will rebuild the `database/migrate_generated.go` file to include execution of all
the packages in `database/migrations`.

## Working with the UIs

FastTrackML incorporates the existing Aim and MLFlow web UIs, albeit with a few
modifications. This is accomplished at build time by importing the
`fasttrackml-ui-aim` and `fasttrackml-ui-mlflow` modules. These repos contain
the patched and compiled UI assets of the upstream repos at specific tagged
revisions. To make a UI change, a PR is merged to the appropriate release branch
and new tag is pushed. At that point, the `fasttrackml` reference can be updated
(in `go.mod`) to pull in the new tag.

For UI development, you'll need a tighter change/view loop, so we recommend the
following approach.

Prerequisites:

- Go 1.20 or higher
- [Docker](https://docs.docker.com/get-docker/)

Steps:

1. Clone the UI repos as siblings to your `fasttrackml` working copy.

    ```bash
    cd ..
    # This is the repo for the Aim UI
    git clone https://github.com/G-Research/fasttrackml-ui-aim.git
    # This is the repo for the MLFlow UI
    git clone https://github.com/G-Research/fasttrackml-ui-mlflow.git
    ```

2. Change the UI working copies to their latest release branch.

    ```bash
    for repo in fasttrackml-ui-*; do
      pushd $repo
      git checkout release/$(cat upstream.txt)
      popd
    done
    ```

3. In the `fasttrackml` folder, use `go work` to map the UI working copy you
   will work on as a go module replacement (we use the Aim UI in this example):

    ```bash
    cd fastrackml
    go work init
    go work use .
    go work use ../fasttrackml-ui-aim
    ```

4. Compile the UI project:

    ```bash
    pushd ../fasttrackml-ui-aim
    go run main.go
    popd
    ```

5. On success, the UI project will have an `embed` directory that contains the
   compiled assets and that can be used directly by the main project. Just build
   and launch it as usual:

    ```bash
    make run
    ```

6. You should now be able to see your local working copy of the UI in your
   browser. As you make changes in the UI's `/src` folder, just re-run the
   compile steps and refresh your browser.

7. When ready, make a PR of your changes to the UI repo, with the merge target
   set as the release branch, _not_ the `main` branch.
