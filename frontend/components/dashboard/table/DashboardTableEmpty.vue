<template>
    <div class="empty delayed-fadein-animation">
      <span class="big_text">{{ $t('dashboard.empty.empty_table') }}</span>
    </div>
  </template>
  
  <style lang="scss" scoped>
  .empty {
    width: 100%;
    height: 400px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: var(--text-color-disabled);
  }
  </style>