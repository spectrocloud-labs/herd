# herd

Herd is a Embedded Runnable DAG (H.E.R.D.). it aims to be a tiny library that allows to define arbitrary DAG, and associate job operations on them.


## Why?

I've found couple of nice libraries ([fx](https://github.com/uber-go/fx), or [dag](https://github.com/mostafa-asg/dag) for instance), however none of them satisfied my constraints:

- Tiny
- Completely tested (TDD)
- Define jobs in a DAG, runs them in sequence, execute the ones that can be done in parallel (parallel topological sorting) in separate go routines
- Provide some sorta of similarity with `systemd` concepts


## Usage

`herd` can be used as such:

```golang


import (
    "github.com/mudler/herd"
)

func main() {
    g := herd.DAG()
    g.Add()
    g.Run(context.TODO())
}

```