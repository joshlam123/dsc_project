convertToTitleCase: function(str) {
     str = str.toLowerCase().split(' ');
     for (var i = 0; i < str.length; i++) {
       str[i] = str[i].charAt(0).toUpperCase() + str[i].slice(1);
     }
     return str.join(' ');
   },

// To format the “possibility” (of weather) string obtained from the weather API
formatPossibility: function(str) {
     str = str.toLowerCase().split('-');
     for (var i = 0; i < str.length; i++) {
       str[i] = str[i].charAt(0).toUpperCase() + str[i].slice(1);
     }
     return str.join(' ');
   },

// To convert Unix timestamps according to our convenience
unixToHuman: function(timezone, timestamp) {
     /* READ THIS BEFORE JUDGING & DEBUGGING
     For any location beyond the arctic circle and the
     antarctic circle, the goddamn weather api does not return certain
     keys/values in each of this.rawWeatherData.daily.data[some_array_index].
     Due to this, console throws up an error.
     The code is correct, the problem is with the API.
     May be later on I will add some padding to tackle missing values.
     */
     var moment = require('moment-timezone'); // for handling date & time
     var decipher = new Date(timestamp * 1000);
     var human = moment(decipher)
       .tz(timezone)
       .format('llll');
     var timeArray = human.split(' ');
     var timeNumeral = timeArray[4];
     var timeSuffix = timeArray[5];
     var justTime = timeNumeral + ' ' + timeSuffix;
     var monthDateArray = human.split(',');
     var monthDate = monthDateArray[1].trim();
     return {
       fullTime: human,
       onlyTime: justTime,
       onlyMonthDate: monthDate
     };
   },

// To convert temperature from fahrenheit to celcius
fahToCel: function(tempInFahrenheit) {
     var tempInCelcius = Math.round((5 / 9) * (tempInFahrenheit — 32));
     return tempInCelcius;
   },

// To convert the air pressure reading from millibar to kilopascal
milibarToKiloPascal: function(pressureInMilibar) {
     var pressureInKPA = pressureInMilibar * 0.1;
     return Math.round(pressureInKPA);
   },

// To convert distance readings from miles to kilometers
mileToKilometer: function(miles) {
     var kilometer = miles * 1.60934;
     return Math.round(kilometer);
   },

// To format the wind direction based on the angle
deriveWindDir: function(windDir) {
     var wind_directions_array = [
       { minVal: 0, maxVal: 30, direction: 'N' },
       { minVal: 31, maxVal: 45, direction: 'NNE' },
       { minVal: 46, maxVal: 75, direction: 'NE' },
       { minVal: 76, maxVal: 90, direction: 'ENE' },
       { minVal: 91, maxVal: 120, direction: 'E' },
       { minVal: 121, maxVal: 135, direction: 'ESE' },
       { minVal: 136, maxVal: 165, direction: 'SE' },
       { minVal: 166, maxVal: 180, direction: 'SSE' },
       { minVal: 181, maxVal: 210, direction: 'S' },
       { minVal: 211, maxVal: 225, direction: 'SSW' },
       { minVal: 226, maxVal: 255, direction: 'SW' },
       { minVal: 256, maxVal: 270, direction: 'WSW' },
       { minVal: 271, maxVal: 300, direction: 'W' },
       { minVal: 301, maxVal: 315, direction: 'WNW' },
       { minVal: 316, maxVal: 345, direction: 'NW' },
       { minVal: 346, maxVal: 360, direction: 'NNW' }
     ];
     var wind_direction = '';
     for (var i = 0; i < wind_directions_array.length; i++) {
       if (
         windDir >= wind_directions_array[i].minVal &&
         windDir <= wind_directions_array[i].maxVal
       ) {
         wind_direction = wind_directions_array[i].direction;
       }
     }
     return wind_direction;
   },
