# About
This folder contains examples of the data used to run Pregel as discussed in thereport. It contains weighted and unweighted graph data with 5, 20, and 50 vertices each.

Each folder ('../data/unweighted/..') and ('../data/weighted/..') contains the following file structure:

├── unweighted
│   ├── prob
│   └── uwgraph.go
└── weighted
    ├── prob
    └── wgraph.go

The folder '../prob' contains the problems in .json format, while the \*graph.go file is used to randomly generate strongly connected graphs. 

# Unweighted Graphs
Unweighted graphs are graphs with unweighted vertices (i.e. vertices do not contain any weight assigned to them.) They are contained in the folder '../unweighted'. The random graph generator will generate integer values for each graph vertice.

# Weighted Graphs
Weighted graphs are graphs with weighted vertices (i.e. vertices have weights assigned to them.) They are contained in the folder '../weighted'. The random graph generator will generate  floating point values for each graph edge, and integer values for each graph vertice.


# Running the Random Graph Generator
To run the random graph generator, change your directory to the type of graph you wish to generate (weighted / unweighted) using *cd weighted* or *cd unweighted*.

When in the folder, the format for generating a weighted graph is as follows:
*go run wgraph.go <name of graph> <number of nodes>*. So for generating a 50 node weighted graph that is named 'random', use the following command:

*go run wgraph.go rand 50*

The same applies for unweighted graphs. 
