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

## Filling the database

It's often necessary to test out your changes on a loaded database, and we definitely want to do this
before making a release. A production-level target for database shape is:
* 40k runs across 300 experiments
* 2k unique metrics per run, each with 200 values

To get some percentage of this into your local database instance, you can use the included K6
load test script, which can be run in the dev container.

1. Start the FML tracking server
2. Start a terminal in the dev container. 
3. `cd docs/example`
4. `k6 run k6_load.js`

Tweak the numbers in `k6_load.js` for number of runs, metrics, etc -- the default amounts are 
pretty small.

## Working with the UIs

FastTrackML incorporates the existing Aim and MLFlow web UIs, albeit
with a few modifications. This is accomplished by importing the
`fasttrackml-ui-aim` and `fasttrackml-ui-mlflow` go modules. The
corresponding repos contain the patched and compiled UI assets of the
upstream repos. To make a UI change, a PR is merged to the appropriate
release branch and a new tag is pushed. At that point, the
`fasttrackml` reference can be updated (in `go.mod`) to pull in the
new tag with your changes.

For UI development, you'll need a tighter change/view loop, so we recommend the
following approach.

Steps:

1. Fetch the UI repo submodules in your working copy of FastTrackML:

    ```bash
    git submodule update --init --recursive
    ```

2. Update the UI submodule to the most recent release branch, and make
   changes as needed:

    ```bash
    cd ui/fasttrackml-ui-aim
    git fetch -a
    git switch release/v3.17.5
    <make edits>
    ```

3. Run the UI development server to see your changes. Make sure the
   FML tracking server is already launched, then run the ui make
   target in the vscode terminal:

    ```bash
    cd <fasttrackml project root>
    make ui-aim-start
    <ctrl-c to stop>
    ```

4. As you make changes, just `sync` them to the development server
   location for fast live update:

    ```bash
    cd <fasttrackml project root>
    make ui-aim-sync
    ```

5. When ready, make a new branch in the submodule, commit changes, and
   push to your fork. Make a PR with the merge target set as the
   release branch, _not_ the `main` branch.
