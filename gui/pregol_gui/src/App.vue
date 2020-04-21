<template>
<div id="ancestor">
   <div class="container-fluid" id="app">
     <div class="row">
       <div id="sidebar" class="col-md-3 col-sm-4 col-xs-12 sidebar">
         <div id="search">
           

             <img src="./assets/pregol.png" width="100" height="48">

         </div>

         <div id="info">

           <div class="wrapper-left pt-4 pb-2 text-center">
           <span>Graph Name:</span>
             <div id="filename">
               <b> {{ graphName }} </b>
             </div>
            </div>

            <div class="wrapper-left pt-4 pb-2 text-center">
            <span>Size of Graph:</span>
               <div id="graph-size"> 
               <b>{{ numVertices }}</b>
               </div>
            </div>

            <div class="wrapper-left pt-4 pb-2 text-center">
            <span>User Defined Function:</span>

             <div id="udf_name">
               <b> {{ 'Max Value' }} </b>
             </div>
            </div>

            <div class="wrapper-left pt-4 pb-2 text-center">
            <span>Number of Partitions:</span>

             <div id="partition_size">
               <b> {{ numPartitions }} </b>
             </div>
            </div>

            <div class="wrapper-left pt-4 pb-2 text-center">
           <span>Current Superstep:</span>
             <div id="currentiteration">
               <b> {{ currentIteration }} </b>
             </div>
            </div>

        </div>
        
        
           
       </div>

       <dashboard-content
         class="col-md-9 col-sm-8 col-md-6 col-xs-12 content"
         id="dashboard-content"
         :highlights="highlights"
         :tempVar="tempVar">

       </dashboard-content>
       </div>
     </div>
   </div>
</template>

<script>
import Content from './components/Content.vue'
import axios from 'axios';

export default {
  name: 'app',
  
  components: {
    'dashboard-content': Content
  },

  data() {
   return {
    list: [],

    graphName:      this.graphName,
    graphFile:      this.graphFile,
    numPartitions:  this.numPartitions,
    numVertices:    this.numVertices,

    currentIteration: this.currentIteration,
    numActiveNodes:  this.numActiveNodes,
    activeNodesVert: this.activeNodesVert,

     
    tempVar: {
      nodeVertCostFn: this.nodeVertCostFn,
      totalAliveTime:  this.totalAliveTime,
       tempToday: [
         // gets added dynamically by this.getSetHourlyTempInfoToday()
       ],
     },

     highlights: {
       details: {
        doneSignal: this.doneSignal,
        numActiveNodes:  this.numActiveNodes,
        avgTiming: this.avgTiming, 
        avgTimingArr: this.avgTiming,
       },
       

       uvIndex: '',
       visibility: '',
       windStatus: {
         windSpeed: '',
         windDirection: '',
         derivedWindDirection: ''
        },
      }
    };
  },
  computed () {

  },

  created () {
        this.fetchData();
    },

  mounted() {
      this.fetchData('app');
  },
  methods: {
  async fetchData ()
   {
    
    var data = {};
    const options = { method: 'GET', 
    url: 'http://127.0.0.1:3000/guiserver', headers: {'Accept': 'application/json', 'Content-Type': 'application/json;charset=UTF-8'}}
    console.log("Fetching data")

    axios(options).then(result => { 
        console.log("Fetched Data")
        console.log(result.status)
        /*eslint-disable*/
         console.log(result.data) 
         /*eslint-enable*/

         // this.response = result.data;
         this.graphName = result.data['GraphName']
         this.graphFile = result.data['GraphFile']
         this.numPartitions = result.data['NumPartitions']
         this.numVertices = result.data['NumVertices']
         this.currentIteration = result.data['CurrentIteration']
         this.avgTiming = result.data['AvgTiming']
        

          if (result.data['DoneSignal'] == 1) {
            var DoneSignal = true
          } else {
            var DoneSignal = false
          }

         this.highlights.details.doneSignal = DoneSignal
         this.highlights.details.numActiveNodes = result.data['NumActiveNodes'].length
         this.highlights.avgTiming = result.data['AvgTiming']
         this.highlights.avgTimingArr = this.highlights.avgTiming[this.highlights.avgTiming.length-1]
         this.tempVar.nodeVertCostFn = result.data['NodeVertCostFn']
         this.tempVar.totalAliveTime = result.data['TotalAliveTime']

       }).catch( error => {
           /*eslint-disable*/
           console.log(error);
           /*eslint-enable*/
     });},

  



  detectEnterKeyPress() {
       var input = this.$refs.input;
       input.addEventListener('keyup', function(event) {
         event.preventDefault();
         var enterKeyCode = 13;
         if (event.keyCode === enterKeyCode) {
           this.setHitEnterKeyTrue();
         }
       });
  },


  },

  

}
</script>


<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

h1, h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}
</style>
