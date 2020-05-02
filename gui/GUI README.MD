# Introduction
This set of folders (../dsc_project/gui/..) contains the source code for the Graphical User Interface built for our implementation of Pregel. 

The entire source code for the graphical user interface was built using Vue.JS. This is managed using the node pacakage manager. Given you have downloaded the and installed the source code from github, these are the first steps you need to take to run the GUI.

**Note: A sample of the save file is provided as guiSave.json in the results folder. Currently, master saves a fresh overwritten copy in that same folder when there is an update to the superstep. So if you run the GUI, you may occasionally see that the file has been refreshed. That is how the GUI refreshes itself, based on the updates to the file.**

# Running the GUI
1) INSTALLATION OF NECEESARY JAVASCRIPT COMPONENTS (VUE & JS RELATED)
cd pregol_gui
npm install # installing the dependencies
npm run dev # to start the GUI

2) LAUNCHING THE GUI
npm run dev # to start the gui

3) LAUNCHING THE GUISERVER
In a separate command line / terminal, go to the same folder and type: go run run_gui.go <n> 
*NOTE: <n> represents the port number on your laptop, i.e. 9000. So it is run this way: go run run_gui.go 9000*

4) UPDATING THE GUI
The GUI automatically fetches new data from the save file whenever the page is refreshed. So refresh the page and you will see requests being made to the server.
