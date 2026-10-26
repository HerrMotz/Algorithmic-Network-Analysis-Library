"""Demo of the Python prototype; run it from the repository root with

    python -m pygraphlib.main
"""
import tempfile
from pathlib import Path

from .graph_io import read_metis_list, read_edge_list, write_edge_list, write_metis_list

EXAMPLES = Path(__file__).resolve().parent.parent / "exampleGraphs"

if __name__ == "__main__":
    # Test 1 
    print("Test 1: \n")
    filepath = EXAMPLES / "metis_tiny_01.mgraph"

    graph = read_metis_list(filepath)
    graph.print_graph()

    # Test 2
    print("Test 2: \n")
    filepath = EXAMPLES / "edge_01.graph"

    graph2 = read_edge_list(filepath)
    graph2.print_graph()


    # Test 3 add_Node, add edge, remove_node, remove_edge
    print("Test 3: \n")
    graph2.add_node(8)
    graph2.add_edge(1, 3)
    graph2.add_edge(7, 8)
    graph2.print_graph()

    graph2.remove_edge(1, 2)
    graph2.remove_node(5)
    graph2.print_graph()

    # ausgeben Nachbar Anz.
    print("Nachbarn von 1: ", graph2.neighbors(1))
    print("Nachbarn von 2: ", graph2.neighbors(2))
    print("Nachbarn von 3: ", graph2.neighbors(3))

    print("Grad von 1: ", graph2.get_node_degree(1))
    print("Grad von 2: ", graph2.get_node_degree(2))
    print("Grad von 3: ", graph2.get_node_degree(3))

    print("1 und 2 sind Nachbarn: ", graph2.is_neighbor(1, 2))
    print("3 und 4 sind Nachbarn: ", graph2.is_neighbor(3, 4))

    # speichern des Ergebnisses von graph2 als METIS und Kantenliste
    # (in einem temporären Verzeichnis, damit die Beispielgraphen unverändert bleiben)
    out = Path(tempfile.mkdtemp(prefix="pygraphlib-"))
    write_metis_list(out / "metis_edge_01_out.mgraph", graph2)
    write_edge_list(out / "edge_01_out.graph", graph2)
    print("Graph gespeichert in", out)