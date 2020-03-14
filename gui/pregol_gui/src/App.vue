<template>
  <div id="app">
    <dashboard-content :highlights="highlights" :tempVar="tempVar"></dashboard-content>
    <Content v-bind:weather_data="weather_data"></Content>
  </div>

  <div id="search">
           <input
             id="location-input"
             type="text"
             ref="input"
             placeholder="Location?"
             @keyup.enter="organizeAllDetails"
           >
           <button id="search-btn" @click="organizeAllDetails">
             <img src="./assets/Search.svg" width="24" height="24">
           </button>
  </div>

<div id="info">
  <div class="wrapper-left">
    <div id="current-weather">
      {{ currentWeather.temp }}
      <span>°C</span>
    </div>
    <div id="weather-desc">{{ currentWeather.summary }}</div>
    <div class="temp-max-min">
      <div class="max-desc">
        <div id="max-detail">
          <i>▲</i>
          {{ currentWeather.todayHighLow.todayTempHigh }}
          <span>°C</span>
        </div>
        <div id="max-summary">at {{ currentWeather.todayHighLow.todayTempHighTime }}</div>
      </div>
      <div class="min-desc">
        <div id="min-detail">
          <i>▼</i>
          {{ currentWeather.todayHighLow.todayTempLow }}
          <span>°C</span>
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

</template>

<script>
import Content from './components/Content.vue'
import * from './utility.js'

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
