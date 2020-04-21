<template>
 <div class="custom-card header-card card">
   <div class="card-body pt-0">
   <div class="wrapper-left pt-4 pb-2 text-center">
     <fusioncharts
       type="column2d"
       width="100%"
       height="250"
       dataformat="json"       

       :datasource="aliveTime"
     > </fusioncharts>
     </div>

     <div class="wrapper-left pt-4 pb-2 text-center">
     </fusioncharts>
     <fusioncharts
       type="msline"
       width="100%"
       height="250"
       dataformat="json"       
       :datasource="cfdata"
     >
     </fusioncharts>
     </div>
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
         caption: "Alive Time for each Node (Supersteps)",
         theme: "umber",
         captionFontBold: "1",
         captionPadding: "10",
         baseFont: "Roboto",
         chartTopMargin: "5",
         showHoverEffect: "1",
         showaxislines: "1",
         numberSuffix: "",
         drawCrossLine: "1",
         plotToolText: "Node: <b>$label</b><br> Alive Time: <b>$dataValue</b>",
         showAxisLines: "0",
         showYAxisValues: "1",
         yaxisname: "No. of Supersteps",
         xaxisname: "Node",
         anchorRadius: "4",
         divLineAlpha: "0",
         labelFontSize: "10",
         labelAlpha: "65",
         labelFontBold: "1",
         rotateLabels: "0",
         slantLabels: "1",
         canvasPadding: "10",
       },
       data: '',
       categories: '',
     },

     cfdata: {
       chart: {
         caption: "Cost Function of each Vertice",
         theme: "fusion",
         captionFontBold: "1",
         captionPadding: "20",
         baseFont: "Roboto",
         chartTopMargin: "15",
         showHoverEffect: "1",
         showaxislines: "1",
         numberSuffix: "",
         drawCrossLine: "1",
         plotToolText: "Node: <b>$label</b><br> Value: <b>$dataValue</b>",
         showAxisLines: "0",
         showYAxisValues: "1",
         yaxisname: "Cost Function",
         xaxisname: "Node",
         anchorRadius: "4",
         divLineAlpha: "0",
         labelFontSize: "13",
         labelAlpha: "65",
         labelFontBold: "1",
         rotateLabels: "1",
         slantLabels: "1",
         canvasPadding: "20",
       },
       dataset: '',
       categories: '',
     },
   };
 },

 methods: {
  randomColor: function() {
  var letters = '0123456789ABCDEF';
  var color = '#';
  for (var i = 0; i < 6; i++) {
    color += letters[Math.floor(Math.random() * 16)];
    }
    return color;
  },

   setcfdata: function() {

     var data = [];

     var allSeriesData = new Object();
     console.log(this.tempVar.nodeVertCostFn)

     var lastThree = [];
     for (var k of Object.keys(this.tempVar.nodeVertCostFn)) {
        lastThree.push(k)
     }
     var last = lastThree.slice(lastThree.length-3, lastThree.length);
     console.log(last)

     for (var k of Object.keys(this.tempVar.nodeVertCostFn)) {
      if (lastThree.indexOf(k) != -1) {
        var allSeriesData = new Object()
          allSeriesData['seriesname'] = k
          allSeriesData['data'] = [];

          for (var key of Object.keys(this.tempVar.nodeVertCostFn[k])) {
            var dataObject = {
              value: this.tempVar.nodeVertCostFn[k][key],
            };
            allSeriesData.data.push(dataObject);
           }

         data.push(allSeriesData)
        }
     }

     this.cfdata.dataset = data;
    console.log(this.cfdata.dataset)
     var category = [];

     var allSeriesData = new Object();
     allSeriesData['category'] = [];
     for (var k of Object.keys(this.tempVar.nodeVertCostFn[1])) {
         var dataObject = {
           label: k,
        }
        allSeriesData.category.push(dataObject);
     }
     category.push(allSeriesData)

     this.cfdata.categories = category;

     console.log(this.cfdata.categories)
   },

   setalivedata: function() {
     var data = [];
     for (var key of Object.keys(this.tempVar.totalAliveTime)) {
       var dataObject = {
         label: key.substr(key.length - 4),
         value: this.tempVar.totalAliveTime[key],
       };
       data.push(dataObject);
     }

     this.aliveTime.data = data;
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
