# Introduction \& General
This is the main README file for the project by Team Pregel for the 50.041 Distribued Systems and Computing project. There are separate readme files in the different directories to explain. 

The general structure of the entire project is as follows: 
.<br>
├── LICENSE <br>
├── README.MD <br>
├── examples <br>
├── gui <br>
├── run_master <br>
├── run_worker <br>
└── src <br>


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

# To Run 
To run the program with n worker nodes and 2 masters (1 primary and 1 secondary), you will need to open n + 2 terminals concurrently. 

For each worker:
1. Ensure working directory is in ```dsc_project/run_worker```
2. ```go run run_worker.go <port#>``` - port# for each worker has to be unique. 

For Primary master
1. Ensure working directory is in ```dsc_project/run_master```
2. ```go run run_master.go <port#> [port# of primary]```

The program is set to run the Max Value example on a 20 node graph. You can change these parameters at ```run_master.go```

