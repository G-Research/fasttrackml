# Maintainer guide

## Release workflow

* All development goes into the `main` branch via pull requests that need to be
  approved and checks that need to pass.
* When we want to cut a new major or minor release, we can create a new release
  branch named `release/MAJOR.MINOR` (e.g. `release/0.4`) and the accompanying
  tag by manually running the `Create release` workflow, specifying the complete
  semantic version (e.g. `0.4.0` or `0.4.0-beta.1`).
* Development continues on the `main` branch.
* Fixes can be backported to the release branch via pull requests that need to
  be approved and checks that need to pass.
* A new patch release can then be made by calling the `Create release` workflow
  again.

## How is it enforced

* A GitHub app needs to exist and be installed on the repo with the
  `contents:write` permissions. Its ID and private key need to be stored as
  secrets under the `create-release` environment. This environment needs to be
  limited to the `main` branch only, and require approval by the maintainers
  (denying self-reviews).
* An environment named `release` needs to exist and be limited to the `main`
  branch and the `v*` tags only. It needs to contain the Docker Hub credentials
  as secrets and be registered on PyPI as a trusted provider for our Python
  package.
* The following rulesets must exist on the repo to enforce this workflow and
  guarantee that a single maintainer cannot make changes to the codebase that
  have not been reviewed by their peers. They replace any other branch
  protection rules that may have existed before.
  ```json
  [
    {
      "name": "Allow bot to create release branches",
      "target": "branch",
      "conditions": {
        "ref_name": {
          "exclude": [],
          "include": [
            "refs/heads/release/**/*"
          ]
        }
      },
      "rules": [
        {
          "type": "creation"
        }
      ],
      "bypass_actors": [
        {
          "actor_id": 0,
          "actor_type": "Integration",
          "bypass_mode": "always"
        }
      ]
    },
    {
      "name": "Allow bot to create tags",
      "target": "tag",
      "conditions": {
        "ref_name": {
          "exclude": [],
          "include": [
            "~ALL"
          ]
        }
      },
      "rules": [
        {
          "type": "creation"
        }
      ],
      "bypass_actors": [
        {
          "actor_id": 0,
          "actor_type": "Integration",
          "bypass_mode": "always"
        }
      ]
    },
    {
      "name": "Protect all tags",
      "target": "tag",
      "conditions": {
        "ref_name": {
          "exclude": [],
          "include": [
            "~ALL"
          ]
        }
      },
      "rules": [
        {
          "type": "deletion"
        },
        {
          "type": "non_fast_forward"
        },
        {
          "type": "update"
        }
      ],
      "bypass_actors": []
    },
    {
      "name": "Protect main and release branches",
      "target": "branch",
      "conditions": {
        "ref_name": {
          "exclude": [],
          "include": [
            "refs/heads/main",
            "refs/heads/release/**/*"
          ]
        }
      },
      "rules": [
        {
          "type": "deletion"
        },
        {
          "type": "non_fast_forward"
        },
        {
          "type": "required_linear_history"
        },
        {
          "type": "pull_request",
          "parameters": {
            "require_code_owner_review": false,
            "require_last_push_approval": true,
            "dismiss_stale_reviews_on_push": true,
            "required_approving_review_count": 1,
            "required_review_thread_resolution": false
          }
        },
        {
          "type": "required_status_checks",
          "parameters": {
            "required_status_checks": [
              {
                "context": "All required checks succeeded",
                "integration_id": 0
              }
            ],
            "strict_required_status_checks_policy": true
          }
        }
      ],
      "bypass_actors": []
    }
  ]
  ```