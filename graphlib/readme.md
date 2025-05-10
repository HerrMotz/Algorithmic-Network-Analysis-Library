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
- 