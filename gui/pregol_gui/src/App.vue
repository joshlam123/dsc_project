<template>
<div id="ancestor">
   <div class="container-fluid" id="app">
     <div class="row">
       <div id="sidebar" class="col-md-3 col-sm-4 col-xs-12 sidebar">
         <div id="search">
           

        <button id="search-btn" @click="organizeAllDetails">
             <img src="./assets/Search.svg" width="24" height="24">
           </button>
         </div>
         <div id="info">

           <div class="wrapper-left pt-4 pb-2 text-center">
           <span>Graph Name:</span>
             <div id="current-weather">
               {{ name }}
             </div>

            <span>User Defined Function:</span>

             <div id="current-weather">
               {{ funcName }}
             </div>

           <span>Size of Graph:</span>
                 
               </div>
               <div id="min-summary"> {{ graphSize }}</div>
             </div>

           <span>Nodes Processed:</span>

             <div id="current-weather">
               {{ currentProgress.nodeProgress }}
             </div>

             <div id="weather-desc">{{ currentProgress.summary }}</div>
             <div class="temp-max-min">
               <div class="max-desc">
               <span># Active Nodes:</span>
               {{ currentProgress.activeNode }}
               </div>

               <div class="min-desc">
                 <div id="min-detail">
                 <span>Number of Partitions:</span>
                 </div>
                 <div id="min-summary"> {{ currentProgress.numPartitions }} </div>
               </div>


           <div class="wrapper-right">

           &nbsp;
           <div>
           <button id="search-btn" @click="organizeAllDetails">
             Refresh 
           </button>

           </div>
            </div>

           </div>
           
       </div>


       <dashboard-content
         class="col-md-9 col-sm-8 col-xs-12 content"
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

     name: '',
     funcName: '',
     graphSize: '',
     
     currentProgress: {
       nodeProgress: '',
       activeNode: '',
       numPartitions: '',


     },
     tempVar: {
       tempToday: [
         // gets added dynamically by this.getSetHourlyTempInfoToday()
       ],
     },
     highlights: {
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

  methods: {
    makeInputEmpty: function() {
     this.$refs.input.value = '';
   },

makeTempVarTodayEmpty: function() {
     this.tempVar.tempToday = [];
   },

detectEnterKeyPress: function() {
     var input = this.$refs.input;
     input.addEventListener('keyup', function(event) {
       event.preventDefault();
       var enterKeyCode = 13;
       if (event.keyCode === enterKeyCode) {
         this.setHitEnterKeyTrue();
       }
     });
   },

locationEntered: function() {
     var input = this.$refs.input;
     if (input.value === '') {
       this.location = "New York";
     } else {
       this.location = this.convertToTitleCase(input.value);
     }
     this.makeInputEmpty();
     this.makeTempVarTodayEmpty();
   },

  },
  computed: {

  },
}
</script>

<style>

</style>


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
