<template>
 <div class="highlights-item col-md-5 col-sm-6 col-xs-12 border-top" layout-align="center center">
   <div flex="15" layout-align="center center">
     <fusioncharts
       :type="type"
       :width="width"
       :height="height"
       :containerbackgroundopacity="containerbackgroundopacity"
       dataEmptyMessage="i-https://i.postimg.cc/R0QCk9vV/Rolling-0-9s-99px.gif"
       :dataformat="dataformat"
       :datasource="datasource"
     ></fusioncharts>
   </div>
 </div>
</template>

<script>
export default {
 props: ["highlights"],
 components: {},
 data() {
   return {
     type: "angulargauge",
     width: "100%",
     height: "100%",
     containerbackgroundopacity: 0,
     dataformat: "json",
     datasource: {
       chart: {
         caption: "Number of Active Vertices",
         captionFontBold: "0",
         captionFontColor: "#000000",
         captionPadding: "30",
         lowerLimit: "0",
         upperLimit: "",
         lowerLimitDisplay: "0",
         upperLimitDisplay: "",
         showValue: "0",
         theme: "fusion",
         baseFont: "Roboto",
         bgAlpha: "0",
         canvasbgAlpha: "0",
         gaugeInnerRadius: "75",
         gaugeOuterRadius: "110",
         pivotRadius: "0",
         pivotFillAlpha: "0",
         valueFontSize: "20",
         valueFontColor: "#000000",
         valueFontBold: "1",
         tickValueDistance: "2",
         autoAlignTickValues: "1",
         majorTMAlpha: "20",
         chartTopMargin: "30",
         chartBottomMargin: "60"
       },
       colorrange: {
         color: [
           {
             minvalue: "0",
             maxvalue: 0,
             code: "#7DA9E0"
           },
           {
             minvalue: 0,
             maxvalue: "",
             code: "#D8EDFF"
           }
         ]
       },
       annotations: {
         groups: [
           {
             items: [
               {
                 id: "val-label",
                 type: "text",
                 text: "",
                 fontSize: "20",
                 font: "Roboto",
                 fontBold: "1",
                 fillcolor: "#000000",
                 x: "$gaugeCenterX",
                 y: "$gaugeCenterY"
               }
             ]
           }
         ]
       },
       dials: {
         dial: [
           {
             value: "",
             baseWidth: "10",
             radius: "100",
             borderThickness: "1",
             baseRadius: "1",
           }
         ]
       }
     }
   };
 },
 methods: {
     setdialProperty: function() {
          
          console.log(Object.values(this.highlights.activeNodesVert));
          var numActiveVerts = 1;

          for (var key in this.highlights.activeNodesVert) {
            for (var k in this.highlights.activeNodesVert[key]){
              for (var m in this.highlights.activeNodesVert[key][k]) {
                numActiveVerts += this.highlights.activeNodesVert[key][k].length
              }
            }
          }

          this.datasource.chart.upperLimit = this.highlights.numVertices;
          this.datasource.chart.upperLimitDisplay = this.highlights.numVertices;
          this.datasource.annotations.groups[0].items[0].text = numActiveVerts;
          this.datasource.dials.dial[0].value = numActiveVerts;
          this.datasource.colorrange.color[0].maxvalue = numActiveVerts;
          this.datasource.colorrange.color[1].minvalue = "0";
          console.log(numActiveVerts);
     },
  
  },
  mounted: function() {
    this.setdialProperty();
  },
  watch: {
   highlights: {
     handler: function() {
       this.setdialProperty();                             
     },
     deep: true
   },
  },
 
};
</script>