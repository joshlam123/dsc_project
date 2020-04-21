<template>
 <div class="custom-card header-card card">
   <div class="card-body pt-0">
     <fusioncharts
       type="spline"
       width="100%"
       height="100%"
       dataformat="json"       dataEmptyMessage="i-https://i.postimg.cc/R0QCk9vV/Rolling-0-9s-99px.gif"

       :datasource="aliveTime"
     >
     </fusioncharts>
     <fusioncharts
       type="spline"
       width="100%"
       height="100%"
       dataformat="json"       dataEmptyMessage="i-https://i.postimg.cc/R0QCk9vV/Rolling-0-9s-99px.gif"
       :datasource="cfdata"
     >
     </fusioncharts>
   </div>

 </div>
</template>

<script>
export default {
 props: ["tempVar"],
 components: {},
 data() {
   return {

     aliveTime: {
       chart: {
         caption: "Alive Time for each Node",
         captionFontBold: "0",
         captionFontColor: "#000000",
         captionPadding: "30",
         baseFont: "Roboto",
         chartTopMargin: "30",
         showHoverEffect: "1",
         theme: "fusion",
         showaxislines: "1",
         numberSuffix: "°C",
         anchorBgColor: "#6297d9",
         paletteColors: "#6297d9",
         drawCrossLine: "1",
         plotToolText: "$label<br><hr><b>$dataValue</b>",
         showAxisLines: "0",
         showYAxisValues: "0",
         anchorRadius: "4",
         divLineAlpha: "0",
         labelFontSize: "13",
         labelAlpha: "65",
         labelFontBold: "0",
         rotateLabels: "1",
         slantLabels: "1",
         canvasPadding: "20"
       },
       data: [],
     },

     cfdata: {
       chart: {
         caption: "Cost Function of each Vertice",
         captionFontBold: "0",
         captionFontColor: "#000000",
         captionPadding: "30",
         baseFont: "Roboto",
         chartTopMargin: "30",
         showHoverEffect: "1",
         theme: "fusion",
         showaxislines: "1",
         numberSuffix: "°C",
         anchorBgColor: "#6297d9",
         paletteColors: "#6297d9",
         drawCrossLine: "1",
         plotToolText: "$label<br><hr><b>$dataValue</b>",
         showAxisLines: "0",
         showYAxisValues: "0",
         anchorRadius: "4",
         divLineAlpha: "0",
         labelFontSize: "13",
         labelAlpha: "65",
         labelFontBold: "0",
         rotateLabels: "1",
         slantLabels: "1",
         canvasPadding: "20"
       },
       data: [],
     },
   };
 },

 methods: {
   setcfdata: function() {
   console.log(this.tempVar.nodeVertCostFn)
     var data = [];
     for (var key of Object.keys(this.tempVar.nodeVertCostFn)) {
       var dataObject = {
         label: key,
         value: this.tempVar.nodeVertCostFn[key],
       };
       data.push(dataObject);
     }
     this.cfdata.data = data;
     console.log(this.cfdata.data)
   },
   setalivedata: function() {
   console.log(this.tempVar.totalAliveTime)
     var data = [];
     for (var i = 0; i < this.tempVar.tempToday.length; i++) {
       var dataObject = {
         label: this.tempVar.tempToday[i].hour,
         value: this.tempVar.tempToday[i].temp
       };

       data.push(dataObject);
     }
     this.cfdata.data = data;
   },
 },
 mounted: function() {
   this.setcfdata();
   this.setalivedata();
 },
 watch: {
   tempVar: {
     handler: function() {
       this.setcfdata();  
       this.setalivedata()                                 
     },
     deep: true
   },
 },
};
</script>

<style>
</style>
