# PR Pick Action

This GitHub Action helps you to pick some PRs to another branch with creating new PR.
If your project has a branch strategy, and developer needs to bring some PR's merge commit to another branch,
this action helps to create merge PR.

## Inputs

see https://github.com/nakatamixi/pr-pick-action/blob/main/action.yaml

## Usage

```
on:
  workflow_dispatch:
    inputs:
      prs:
        required: true
        description: ""
      to:
        required: true
        description: ""
      base:
        required: false
        description: ""
        default: "main"

jobs:
  test_job:
    runs-on: ubuntu-latest
    name: test workflow
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - id: foo
        uses: nakatamixi/pr-pick-action@v1
        with:
          to: ${{github.event.inputs.to}}
          base: ${{github.event.inputs.base}}
          prs: ${{github.event.inputs.prs}}
          token: ${{ secrets.GITHUB_TOKEN }}
```
