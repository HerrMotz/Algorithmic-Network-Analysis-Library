# Example graphs

Most graphs in this directory come from public network collections. They are
redistributed under the licenses of those collections, which are listed below;
the license of this repository's code does not apply to them. If you use one of
the graphs, please cite its original source.

Vertex and edge counts are those given by the sources. They include self-loops,
which only occur in the Facebook networks (295 and 76) and the road networks
(1 to 20 each).

## Social, collaboration and animal networks

| File | Vertices | Edges | Network | Source | License |
|---|---:|---:|---|---|---|
| `soc-karate.mtx` | 34 | 78 | Friendships in Zachary's karate club | [SuiteSparse Matrix Collection][suitesparse], `Newman/karate` (also on [Network Repository][nr] as `soc-karate`). W. W. Zachary, *An information flow model for conflict and fission in small groups*, J. Anthropological Research 33, 452–473 (1977) | [CC BY 4.0][cc-by-4] |
| `soc-dolphins.mtx` | 62 | 159 | Associations between bottlenose dolphins in Doubtful Sound, New Zealand | [SuiteSparse Matrix Collection][suitesparse], `Newman/dolphins` (also on [Network Repository][nr] as `soc-dolphins`). D. Lusseau et al., Behavioral Ecology and Sociobiology 54, 396–405 (2003) | [CC BY 4.0][cc-by-4] |
| `out.dimacs10-netscie.graph` | 1,461 | 2,742 | Co-authorships among network scientists | [KONECT][konect] `dimacs10-netscience`, from the DIMACS10 clustering archive. M. E. J. Newman, *Finding community structure in networks using the eigenvectors of matrices*, Phys. Rev. E 74(3) (2006) | not stated by KONECT |
| `out.pajek-erdos.graph` | 6,927 | 11,850 | Erdős collaboration network (2002) | [KONECT][konect] `pajek-erdos`, from V. Batagelj's [Pajek datasets][pajek] | not stated by KONECT |
| `fb-pages-media.edges` | 27,917 | 206,259 | Mutual likes between verified Facebook pages of news sites (2017) | [Network Repository][nr] `fb-pages-media`, the "New Sites" network of the [GEMSEC Facebook data][gemsec]. B. Rozemberczki, R. Davies, R. Sarkar, C. Sutton, *GEMSEC: Graph Embedding with Self Clustering*, ASONAM 2019 | [CC BY-SA][nr-policy] |
| `fb-pages-public-figure.edges` | 11,565 | 67,114 | The same for pages of public figures | [Network Repository][nr] `fb-pages-public-figure`, the "Public Figures" network of the [GEMSEC Facebook data][gemsec], see above | [CC BY-SA][nr-policy] |
| `baboon21.edges` | 8 | 15 | Associations in a group of baboons, Amboseli National Park, Kenya | [Network Repository][nr] `mammalia-baboon-association-group21`, from the [Animal Social Network Repository][asnr]. M. Franz, J. Altmann, S. C. Alberts, *Knockouts of high-ranking males have limited impact on baboon social networks*, Current Zoology 61(1), 107–113 (2015) | [CC BY-SA][nr-policy] |
| `fish-guppy-familiar-5.edges` | 6 | 15 | Interactions in a shoal of familiar guppies | [Network Repository][nr] `fish-guppy-familiar-5`, from the [Animal Social Network Repository][asnr]. M. J. Hasenjager, L. A. Dugatkin, *Familiarity affects network structure and information flow in guppy (Poecilia reticulata) shoals*, Behavioral Ecology (2016) | [CC BY-SA][nr-policy] |
| `abi.edgelist` | 63 | 966 | A graph of students in an Abitur class. There is an edge if they share a course. | Maximilian T. Stock | All rights reserved. |

KONECT asks to be cited as J. Kunegis, *KONECT – The Koblenz Network
Collection*, Proc. Int. Conf. on World Wide Web Companion, 1343–1350 (2013), and
Network Repository as R. A. Rossi, N. K. Ahmed, *The Network Data Repository
with Interactive Graph Analytics and Visualization*, AAAI 2015. The SuiteSparse
Matrix Collection asks for T. A. Davis, Y. Hu, *The University of Florida Sparse
Matrix Collection*, ACM Transactions on Mathematical Software 38(1) (2011).

## Road networks

| File | Vertices | Edges |
|---|---:|---:|
| `jena_road_unfolded_50.metis` | 8,747 | 9,485 |
| `jena_road_unfolded_50.edgelist` | 8,747 | 9,485 |
| `erfurt_road_unfolded_50.metis` | 18,618 | 20,130 |
| `weimar_road_unfolded_50.metis` | 5,385 | 5,874 |
| `apolda_road_unfolded_50.metis` | 2,342 | 2,530 |

Drivable road networks of four towns in Thuringia, downloaded from
[OpenStreetMap][osm] with [OSMnx][osmnx] in May 2025. To approximate road
lengths in an unweighted graph, every road segment was subdivided into a path of
⌈length / 50 m⌉ edges ("unfolded_50"). The edge list is a conversion of the Jena
METIS file.

Map data © OpenStreetMap contributors, available under the
[Open Database License (ODbL) 1.0][odbl]. As derived databases, these files are
licensed under the ODbL as well.

## METIS sample graphs

| File | Vertices | Edges | Graph |
|---|---:|---:|---|
| `4elt.mgraph` | 15,606 | 45,878 | 2D finite element mesh |
| `metis_test.mgraph` | 766 | 1,314 | Small graph with two vertex weights per vertex |
| `metis_tiny_01.mgraph` | 7 | 11 | Tiny example graph |
| `tiny_01.graph` | 7 | 11 | `metis_tiny_01.mgraph` as an edge list |

Sample graphs of [METIS][metis] (G. Karypis, V. Kumar), as provided on John
Burkardt's [METIS graph page][burkardt] under the GNU LGPL.

## Small test graphs

`edge_01.graph` (5 vertices, 6 edges) and `graph_for_test.graph` (15 vertices,
50 edges) were made for the tests of this project.

[suitesparse]: https://sparse.tamu.edu
[nr]: https://networkrepository.com
[nr-policy]: https://networkrepository.com/policy.php
[konect]: http://konect.cc
[pajek]: http://vlado.fmf.uni-lj.si/pub/networks/data/
[gemsec]: https://snap.stanford.edu/data/gemsec-Facebook.html
[asnr]: https://bansallab.github.io/asnr/
[cc-by-4]: https://creativecommons.org/licenses/by/4.0/
[osm]: https://www.openstreetmap.org
[osmnx]: https://github.com/gboeing/osmnx
[odbl]: https://opendatacommons.org/licenses/odbl/1-0/
[metis]: https://github.com/KarypisLab/METIS
[burkardt]: https://people.math.sc.edu/Burkardt/data/metis_graph/metis_graph.html
