<template>
 <div class="highlights-item col-md-7 col-sm-5 col-xs-12 border-top">
   <div>
   <div class="card-heading pt-4">Processing Summary</div>
   <div class="row pt-4 mt-4">
     <div class="col-sm-6 col-md-6 mt-2 text-center align-middle">
       <p class="card-sub-heading mt-3">Number of Active Nodes</p>
       <p class="mt-4"><img src="../assets/active.png" height="40" width="40"></p>
       <p class="card-value mt-4"> <b> {{ highlights.details.numActiveNodes }} </b> </p>
     </div>
     <div class="col-sm-6 col-md-6 mt-2">
       <p class="card-sub-heading mt-3">Is Pregel Done?</p>
       <p class="mt-4"><img src="../assets/done.svg" height="40" width="40"></p>
       <p class="card-value mt-4"> <b> {{ highlights.details.doneSignal }} </b> </p>
     </div>
     <div class="col-sm-6 col-md-6 mt-2">
       <p class="card-sub-heading mt-3">Average Time of All Supersteps</p>
       <p class="mt-4"><img src="../assets/windspeed.svg" height="40" width="40"></p>
       <p class="card-value mt-4"> <b> {{ highlights.avgTimingArr }} </b> </p>
     </div>
   </div>
   </div>
 </div>
</template>

<script>
export default {
 props: ["highlights"],
 components: {},
 data() {
   return {};
 },
 methods: {
 },
 computed: {}
};
</script>
