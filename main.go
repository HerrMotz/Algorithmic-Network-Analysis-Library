// Command example demonstrates how to create and populate unweighted,
// undirected graphs with the helpers provided by the graphlib package.
//
// To run it from the project root:
//     go run ./example

package main

import (
	"fmt"
	"log"
	
	"graphlib/graphlib"
)

func main() {
	// -----------------------------
	//  Dense adjacency matrix demo
	// -----------------------------
	mat, err := graphlib.InitializeAdjMat(4)
	if err != nil {
		panic(err)
	}
	mat.AddEdge(0, 1)
	mat.AddEdge(1, 2)

	fmt.Println("Adjacency matrix contains edge (0,1):", mat.HasEdge(0, 1))
	fmt.Println("Adjacency matrix contains edge (2,3):", mat.HasEdge(2, 3))

	// ----------------------------------------
	//  Compressed Sparse Row (CSR) matrix demo
	// ----------------------------------------
	csr, _ := graphlib.InitializeCSRAdjMat(4)
	csr.AddEdge(0, 3)
	csr.AddEdge(0, 2)

	fmt.Println("Neighbours of vertex 0 in CSR:", csr.Neighbours(0))

	// --------------------------
	//  Adjacency list structure
	// --------------------------
	adj, _ := graphlib.InitializeAdjList(4)
	adj.AddArc(1, 3)

	// The adjacency list currently keeps the internal slice unexported; we
	// therefore just demonstrate adding an edge. (You could expose a
	// Neighbours() accessor exactly like the CSR version if you need one.)
	fmt.Println("Added undirected edge (1,3) to adjacency list – done!")

	// ---------------------------------
	//  External name → index lookup map
	// ---------------------------------
	names := []string{"berlin", "hamburg", "munich", "koeln"}
	lut, _ := graphlib.InitializeLookupTable(len(names), names)
	fmt.Println("Index of munich is:", lut["munich"])
}
