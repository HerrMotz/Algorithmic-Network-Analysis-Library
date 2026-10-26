from .graph import Graph


# === Import Functions ===

def read_edge_list(filepath):
    """
    Liest eine Kantenliste ein. Ignoriert Zeilen, die mit '%' oder '#' beginnen.
    Gibt ein Graph-Objekt zurück.
    """
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
            # print(f"Adding edge: {u} - {v}")
            g.add_edge(u, v)

    return g


def read_metis_list(filepath):
    """
    Liest eine METIS-Datei (ohne Gewichte) ein. Ignoriert Zeilen, die mit '%'
    oder '#' beginnen. Eine leere Zeile ist ein Knoten ohne Nachbarn.
    Gibt ein Graph-Objekt zurück.
    """
    g = Graph()
    current_node = 1
    num_nodes = None

    with open(filepath, 'r') as f:
        for line in f:
            line = line.strip()
            if line.startswith('%') or line.startswith('#'):
                continue

            if num_nodes is None:
                if line:
                    num_nodes = int(line.split()[0])  # Header-Zeile "n m ..."
                continue

            if current_node > num_nodes:
                break  # leere Zeilen am Dateiende

            g.add_node(current_node)
            neighbors = map(int, line.split())
            for neighbor in neighbors:
                # print(f"Adding edge: {current_node} - {neighbor}")
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
        for u in g.node_to_index:
            for v in g.neighbors(u, convert_to_original_index=True):
                edge = tuple(sorted((u, v)))
                if edge not in written_edges:
                    f.write(f"{edge[0]} {edge[1]}\n")
                    written_edges.add(edge)


def write_metis_list(filepath, g, directed=False):
    """
    Schreibt den Graphen im METIS-Format mit Header-Zeile.
    Jede nachfolgende Zeile listet die Nachbarn eines Knotens. METIS nummeriert
    die Knoten nach ihrer Position 1..n, nicht nach ihrem Namen.
    """
    num_nodes = g.get_num_nodes()
    num_edges = g.get_num_edges()

    with open(filepath, 'w') as f:
        f.write(f"{num_nodes} {num_edges}\n")
        for index in range(num_nodes):
            neighbor_str = " ".join(str(i + 1) for i in g.neighbors(index))
            f.write(neighbor_str + "\n")
