<script setup>
// App.vue 只做一件事：渲染当前路由。
// 布局不写在这里——前台页面套 DefaultLayout，后台页面套 AdminLayout，
// 由路由表的嵌套结构决定（见 router/index.js）。
</script>

<template>
  <RouterView />
</template>
