#include <stdio.h>
#include <string>

// initializes a symmetric adjacency matrix
char* initialize_adj_mat( size_t size );

// initializes a sparse adjacency matrix
char* initialize_csr_adj_mat ( size_t size );

// initializes an adjacency list of a given size
char* initialize_adj_list ( size_t size );

// initializes a mapping from external names to the internal representation
char* initialize_lookup_table ( size_t size, string* external_names );


