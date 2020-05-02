# Introduction \& General
This is the main README file for the project by Team Pregel for the 50.041 Distribued Systems and Computing project. There are separate readme files in the different directories to explain. 

The general structure of the entire project is as follows: 
.
├── LICENSE
├── README.MD
├── examples
│   ├── data
│   │   ├── unweighted
│   │   └── weighted
│   ├── max_value
│   │   └── max_value.go
│   ├── pagerank
│   │   └── pagerank.go
│   ├── shortest_path
│   │   └── shortest_path.go
│   └── testGraph.go
├── gui
│   ├── GUI\ README.MD
│   ├── pregol_gui
│   │   ├── README.md
│   │   ├── index.html
│   │   ├── node_modules
│   │   ├── package-lock.json
│   │   ├── package.json
│   │   ├── src
│   │   └── webpack.config.js
│   ├── results
│   │   └── guiSave.json
│   └── run_gui.go
├── run_master
│   ├── checkpoint.json
│   ├── ip_add.txt
│   └── run_master.go
├── run_worker
│   └── run_worker.go
└── src
    └── pregol
        ├── graph_reader.go
        ├── gui_server.go
        ├── master.go
        ├── vertex.go
        └── worker.go


# Source Code
The source code for Pregel is contained within the folder '/src'. It includes 5 different code bases mainly for:
1. Graph Reader : A package written to read in graphs in the format appropriate for our implementation *GraphReader*
2. Gui Server : A package written to run the server that the GUI uses to send **GET** requests to 
3. Master : A package written to contain the Master implementation that interacts with Worker
4. Vertex : A package written to contain the vertices of the entire graph and contain methods which vertices use
5. Worker : A package  written to contain the Worker implementation that receives graphs and performs computation on each partition received


# Examples
The examples folder is designed to help the user do two things: 
1. Generate random graphs or use already existing randomly generated graph data for testing pregel (2 categories: Unweighted - containing graphs with unweighted vertices, Weighted - containing graphs with weighted vertices)
2. Generate user defined functions by writing their own type of function. 3 examples are provided: 1) Max Value, 2) PageRank, and 3) Single Source Shortest Path.


# GUI
The GUI folder is a standalone folder which holds the results from each run of Pregel and also the source code for the GUI. It has a separate README which contains information on how to run the GUI. 


