# A Go library for Graphs

## Supported Methods

### Adjacency List (AdjList)
- Init empty *AdjList*
- Add edges to AdjList
- Read from *METIS* or plain *Edgelist* format to *AdjList*

### Dense Adjacency Matrix (AdjMatDense)
- Init empty *AdjMatDense*
- Add edges
- HasEdge
- HasVertex
- Neighbours
- Degree

### To Do's

- [ ] Export to METIS and edgelist format (for AdjList and AdjMat)
- [ ] Implement the GraphRepresentation type for both AdjList and AdjMat
- [ ] Add test for degree


## Considerations for AdjList Datatype
If I use a list of maps, where the keys are node indices and 
the values are boolean, then the query `hasEdge` can be run in $O(1)$.

```go
package graphlib

type AdjList struct {
    Size int
    List []map[int]bool
}
```

But to get all neighbours, I would have to iterate over the
maps and convert them to a list of vertices. To avoid
this, the method to fetch neighbours of a vertex will return
just the map in the corresponding list entry. This makes
fetching neighbours $O(1)$ as well (which is advantageous for
breadth-first search).