# === Import Functions ===

def read_edge_list(filepath):
    """
    Liest eine Kantenliste ein. Ignoriert Zeilen, die mit '%' oder '#' beginnen.
    Gibt ein Graph-Objekt zurück.
    """
    from graphlib.graph import Graph

    g = Graph()

    with open(filepath, 'r') as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('%') or line.startswith('#'):
                continue

            parts = line.split()
            if len(parts) != 2:
                raise ValueError(f"Ungültige Zeile in Kantenliste: '{line}'")

            u, v = map(int, parts)
            g.add_edge(u, v)

    return g


def read_metis_list(filepath):
    """
    Liest eine METIS-Datei ein. Ignoriert Zeilen, die mit '%' oder '#' beginnen.
    Gibt ein Graph-Objekt zurück.
    """
    from graphlib.graph import Graph

    g = Graph()
    current_node = 1
    header_skipped = False

    with open(filepath, 'r') as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('%') or line.startswith('#'):
                continue

            if not header_skipped:
                header_skipped = True
                continue  # Überspringe Header-Zeile

            neighbors = map(int, line.split())
            for neighbor in neighbors:
                g.add_edge(current_node, neighbor)

            current_node += 1

    return g


# === Export Functions ===

def write_edge_list(filepath, g):
    """
    Schreibt den Graphen in eine Kantenliste. Jede Zeile enthält zwei Knoten.
    Duplikate (z.B. 1-2 und 2-1) werden vermieden.
    """
    written_edges = set()

    with open(filepath, 'w') as f:
        for u in range(1, g.get_num_nodes() + 1):
            for v in g.neighbors(u):
                edge = tuple(sorted((u, v)))
                if edge not in written_edges:
                    f.write(f"{edge[0]} {edge[1]}\n")
                    written_edges.add(edge)


def write_metis_list(filepath, g):
    """
    Schreibt den Graphen im METIS-Format mit Header-Zeile.
    Jede nachfolgende Zeile listet die Nachbarn eines Knotens (1-basiert).
    """
    num_nodes = g.get_num_nodes()
    num_edges = sum(len(g.neighbors(u)) for u in range(1, num_nodes + 1)) // 2

    with open(filepath, 'w') as f:
        f.write(f"{num_nodes} {num_edges}\n")
        for u in range(1, num_nodes + 1):
            neighbors = g.neighbors(u)
            f.write(" ".join(map(str, neighbors)) + "\n")
