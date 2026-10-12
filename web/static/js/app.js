/* ==========================================================================
   LabShop Origin Application Script
   Used for testing CDN Edge Caching and Auto-Minification
   ========================================================================== */

(function () {
  'use strict';

  console.log("LabShop application runtime initialized");

  window.LabShop = {
    version: "1.0.0",
    checkHealth: async function () {
      const response = await fetch("/api/health");
      return await response.json();
    },
    getTime: async function () {
      const response = await fetch("/api/time");
      return await response.json();
    }
  };
})();
