<template>
  <div id="pagecontent">
    <temp-var-chart :tempVar="tempVar"></temp-var-chart>
    <today-highlights :highlights="highlights"></today-highlights>
  </div>
</template>

<script>
import TempVarChart from './TempVarChart.vue'
import Highlights from './Highlights.vue'

export default {
props: ['highlights', 'tempVar'],
components: {
    'temp-var-chart': TempVarChart,
    'today-highlights': Highlights
  },
  data () {
  
  },
  methods: {

  },
  computed: {

  },
}
</script>

<style>
#pagecontent {
  border: 1px solid black;
  padding: 2px;
}
</style>

