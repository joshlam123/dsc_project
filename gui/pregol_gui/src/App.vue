<template>
<div id="ancestor">
   <div class="container-fluid" id="app">
     <div class="row">
       <div id="sidebar" class="col-md-3 col-sm-4 col-xs-12 sidebar">
         <div id="search">
           <input
             id="location-input"
             type="text"
             ref="input"
             placeholder="Location?"
             @keyup.enter="organizeAllDetails">

<button id="search-btn" @click="organizeAllDetails">
             <img src="./assets/Search.svg" width="24" height="24">
           </button>
         </div>
         <div id="info">

           <div class="wrapper-left pt-4 pb-2 text-center">
           <span>Graph Name:</span>
           <i>▼</i>
             <div id="current-weather">
               {{ currentWeather.temp }}
             </div>

            <span>User Defined Function:</span>
           <i>▼</i>
             <div id="current-weather">
               {{ currentWeather.temp }}
             </div>

           <span>Nodes Processed:</span>
           <i>▼</i>
             <div id="current-weather">
               {{ currentWeather.temp }}
             </div>

             <div id="weather-desc">{{ currentWeather.summary }}</div>
             <div class="temp-max-min">
               <div class="max-desc">
               <span># Active Nodes:</span>
               <i>10</i>
               </div>

               <div class="min-desc">
                 <div id="min-detail">
                 <span>Size of Graph:</span>
                   <i>▼</i>
                   {{ currentWeather.todayHighLow.todayTempLow }}
                 </div>
                 <div id="min-summary">at {{ currentWeather.todayHighLow.todayTempLowTime }}</div>
               </div>
             </div>
           </div>
           <div class="wrapper-right">
             <div class="date-time-info">
               <div id="date-desc">
                 <img src="./assets/calendar.svg" width="20" height="20">
                 {{ currentWeather.time }}
               </div>
             </div>
             <div class="location-info">
               <div id="location-desc">
                 <img
                   src="./assets/location.svg"
                   width="10.83"
                   height="15.83"
                   style="opacity: 0.9;"
                 >
                 {{ currentWeather.full_location }}
                 <div id="location-detail" class="mt-1">
                   Lat: {{ currentWeather.formatted_lat }}
                   <br>
                   Long: {{ currentWeather.formatted_long }}
                 </div>
               </div>
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


export default {
  name: 'app',
  components: {
    'dashboard-content': Content
  },
  data() {
   return {
     weatherDetails: false,
     location: '', // raw location from input
     lat: '', // raw latitude from google maps api response
     long: '', // raw longitude from google maps api response
     completeWeatherApi: '', // weather api string with lat and long
     rawWeatherData: '', // raw response from weather api
     currentWeather: {
       full_location: '', // for full address
       formatted_lat: '', // for N/S
       formatted_long: '', // for E/W
       time: '',
       temp: '',
       todayHighLow: {
         todayTempHigh: '',
         todayTempHighTime: '',
         todayTempLow: '',
         todayTempLowTime: ''
       },
       summary: '',
       possibility: ''
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
