<script>
import { authComputed } from '@state/helpers'
import { applySavedTheme, toggleTheme } from '@utils/theme-modern'

export default {
  data() {
    return { theme: 'dark', mobileOpen: false }
  },
  computed: {
    ...authComputed,
    isAdmin() { return this.loggedIn && this.currentUser && this.currentUser.role === 'ADMIN' },
    links() {
      if (!this.loggedIn) return [{ name: 'login', title: 'Anmelden', icon: '↪' }]
      return [
        { name: 'home', title: 'Dashboard', icon: '▦' },
        { name: 'vehicles', title: 'Fahrzeuge', icon: '▰' },
        { name: 'tires', title: 'Reifen', icon: '◉' },
        { name: 'drivers', title: 'Fahrer', icon: '♙' },
        { name: 'recurring-costs', title: 'Fixkosten & Finanzierung', icon: '€' },
        { name: 'maintenance', title: 'Wartungscenter', icon: '⚒' },
        { name: this.isAdmin ? 'backup' : 'import', title: 'Backup & Umzug', icon: '⇄' },
        { name: 'settings', title: 'Einstellungen', icon: '⚙' },
        ...(this.isAdmin ? [
          { name: 'site-settings', title: 'Systemeinstellungen', icon: '☷' },
          { name: 'users', title: 'Benutzer', icon: '♙' },
        ] : []),
        { name: 'logout', title: 'Abmelden', icon: '⇥' },
      ]
    },
  },
  created() { this.theme = applySavedTheme() },
  watch: { '$route'() { this.mobileOpen = false } },
  methods: {
    switchTheme() { this.theme = toggleTheme() },
    closeMenu() { this.mobileOpen = false },
    activeLink(link) {
      const name = this.$route.name || ''
      if (link.name === 'vehicles') return name === 'vehicles' || name.indexOf('vehicle-') === 0
      if (link.name === 'backup' || link.name === 'import') return name === 'backup' || name === 'import' || name.indexOf('import-') === 0
      return name === link.name
    },
  },
}
</script>

<template>
  <div class="hm-navigation">
    <header class="hm-mobile-bar">
      <router-link class="hm-brand" :to="{ name: 'home' }">HAMMOND <span>MODERN</span></router-link>
      <button class="hm-menu-toggle" type="button" :aria-expanded="mobileOpen.toString()" aria-label="Menü öffnen oder schließen" @click="mobileOpen = !mobileOpen">☰</button>
    </header>
    <div v-if="mobileOpen" class="hm-sidebar-backdrop" @click="closeMenu"></div>
    <aside class="hm-sidebar" :class="{ 'hm-sidebar-open': mobileOpen }" aria-label="Hauptnavigation">
      <router-link class="hm-brand hm-desktop-brand" :to="{ name: 'home' }" @click.native="closeMenu">HAMMOND <span>MODERN</span></router-link>
      <p class="hm-nav-heading">WORKSPACE</p>
      <nav class="hm-nav-links">
        <router-link v-for="link in links" :key="link.name" :to="{ name: link.name }" class="hm-nav-link" :class="{ 'hm-nav-active': activeLink(link) }" @click.native="closeMenu">
          <span class="hm-nav-icon" aria-hidden="true">{{ link.icon }}</span>
          <span>{{ link.title }}</span>
          <span v-if="link.badge" class="hm-nav-badge">{{ link.badge }}</span>
        </router-link>
      </nav>
      <div class="hm-sidebar-footer">
        <button class="hm-theme-switch" type="button" :aria-label="theme === 'dark' ? 'Helles Design aktivieren' : 'Dunkles Design aktivieren'" @click="switchTheme">
          <span aria-hidden="true">{{ theme === 'dark' ? '☀' : '☾' }}</span>
          {{ theme === 'dark' ? 'Helles Design' : 'Dunkles Design' }}
        </button>
        <span class="hm-version">Hammond Modern · Community Edition</span>
      </div>
    </aside>
  </div>
</template>
