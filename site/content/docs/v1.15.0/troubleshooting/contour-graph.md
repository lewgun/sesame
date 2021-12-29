# Visualizing Sesame's Internal Object Graph

Sesame models its configuration using a directed acyclic graph (DAG) of internal objects.
This can be visualized through a debug endpoint that outputs the DAG in [DOT][2] format.
To visualize the graph, you must have [`graphviz`][3] installed on your system.

To download the graph and save it as a PNG:

```bash
# Port forward into the sesame pod
$ SESAME_POD=$(kubectl -n projectsesame get pod -l app=sesame -o name | head -1)
# Do the port forward to that pod
$ kubectl -n projectsesame port-forward $SESAME_POD 6060
# Download and store the DAG in png format
$ curl localhost:6060/debug/dag | dot -T png > sesame-dag.png
```

The following is an example of a DAG that maps `http://kuard.local:80/` to the
`kuard` service in the `default` namespace:

![Sample DAG][4]

[2]: https://en.wikipedia.org/wiki/DOT
[3]: https://graphviz.gitlab.io/
[4]: /img/kuard-dag.png
