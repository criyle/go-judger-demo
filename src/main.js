import Vue from 'vue'
import VueMaterial from 'vue-material'
import App from './App.vue'
import * as moment from 'moment'
import 'vue-material/dist/vue-material.min.css'

Vue.use(VueMaterial)

Vue.config.productionTip = false

Vue.filter('date', function(value) {
  const d = new Date(value)
  return moment(d).format('dddd, MMMM Do YYYY, kk:mm:ss.SSS ZZ')
})

Vue.filter('cpu', function(value) {
  if (value) {
    return value + ' ms'
  } else {
    return '0 ms'
  }
})

Vue.filter('memory', function(value) {
  if (value) {
    return value + ' kB'
  } else {
    return '0 kB'
  }
})

new Vue({
  render: h => h(App),
}).$mount('#app')