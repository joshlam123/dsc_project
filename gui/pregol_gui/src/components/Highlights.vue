<template>
  <div class="custom-content-card content-card card">
    <div class="card-body pb-0">
    <div class="content-header h4 text-center pt-2 pb-3">Highlights</div>
      <div class="row">
        <active-nodes :highlights="highlights"></active-nodes>
        <sum-index :highlights="highlights"></sum-index>
      </div>
    </div>
  </div>
</template>

<script>
import ActiveNodes from "./activenodes.vue";
import Details from "./Summary.vue";

export default {
  props: ["highlights"],
  components: {
    "active-nodes": ActiveNodes,
    "sum-index": Details,
  },
};
</script>