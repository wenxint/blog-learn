<script setup>
// App.vue 只负责「装配布局 + 渲染当前路由」。
// 具体页面长什么样，由 views/ 下的组件决定。
import DefaultLayout from './layouts/DefaultLayout.vue'
</script>

<template>
  <DefaultLayout>
    <!-- RouterView 是路由的出口：地址匹配到哪个组件，就渲染在这里 -->
    <RouterView />
  </DefaultLayout>
</template>
