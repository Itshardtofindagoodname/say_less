package web

// BrowserRuntime contains the minimal JavaScript runtime for Say Less Web
const BrowserRuntime = `
// Say Less Web Runtime v0.1.0
// Minimal reactive runtime for browser-side interactivity
(function() {
  'use strict';

  // Runtime state
  const state = {};
  const watchers = {};
  const eventHandlers = {};

  // State management
  function setState(name, value) {
    const old = state[name];
    state[name] = value;
    if (old !== value) {
      notifyWatchers(name, value);
    }
  }

  function getState(name) {
    return state[name];
  }

  function notifyWatchers(name, value) {
    if (watchers[name]) {
      watchers[name].forEach(fn => fn(value));
    }
  }

  function watch(name, fn) {
    if (!watchers[name]) {
      watchers[name] = [];
    }
    watchers[name].push(fn);
  }

  // DOM manipulation helpers
  function $(selector) {
    return document.querySelector(selector);
  }

  function $$(selector) {
    return document.querySelectorAll(selector);
  }

  function createElement(tag, attrs, children) {
    const el = document.createElement(tag);
    for (const [key, val] of Object.entries(attrs || {})) {
      if (key.startsWith('on')) {
        const event = key.slice(2).toLowerCase();
        el.addEventListener(event, val);
      } else {
        el.setAttribute(key, val);
      }
    }
    for (const child of (children || [])) {
      if (typeof child === 'string') {
        el.appendChild(document.createTextNode(child));
      } else if (child) {
        el.appendChild(child);
      }
    }
    return el;
  }

  // Event delegation
  function delegate(parent, event, selector, handler) {
    parent.addEventListener(event, (e) => {
      const target = e.target.closest(selector);
      if (target && parent.contains(target)) {
        handler.call(target, e);
      }
    });
  }

  // Initialize reactive bindings
  function initReactive() {
    document.querySelectorAll('[data-state]').forEach(el => {
      const stateName = el.getAttribute('data-state');
      const template = el.getAttribute('data-template');
      if (template) {
        watch(stateName, () => {
          el.textContent = renderTemplate(template);
        });
      }
    });
  }

  // Template rendering
  function renderTemplate(template) {
    return template.replace(/\$\{([^}]+)\}/g, (match, expr) => {
      try {
        const fn = new Function(...Object.keys(state), 'return ' + expr);
        return fn(...Object.values(state));
      } catch (e) {
        console.warn('Template render error:', e);
        return '';
      }
    });
  }

  // Initialize event handlers
  function initEvents() {
    document.querySelectorAll('[data-click]').forEach(el => {
      const handler = el.getAttribute('data-click');
      el.addEventListener('click', () => executeHandler(handler));
    });

    document.querySelectorAll('[data-input]').forEach(el => {
      const handler = el.getAttribute('data-input');
      el.addEventListener('input', () => executeHandler(handler));
    });

    document.querySelectorAll('[data-submit]').forEach(el => {
      const handler = el.getAttribute('data-submit');
      el.addEventListener('submit', (e) => {
        e.preventDefault();
        executeHandler(handler);
      });
    });
  }

  function executeHandler(code) {
    try {
      const fn = new Function(...Object.keys(state), code);
      fn(...Object.values(state));
    } catch (e) {
      console.error('Handler error:', e);
    }
  }

  // Fetch helper
  async function fetchData(url, options) {
    try {
      const response = await fetch(url, options);
      if (!response.ok) {
        throw new Error('HTTP ' + response.status);
      }
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }
      return await response.text();
    } catch (e) {
      console.error('Fetch error:', e);
      throw e;
    }
  }

  // Storage helpers
  function saveLocal(key, value) {
    localStorage.setItem(key, JSON.stringify(value));
  }

  function loadLocal(key) {
    const val = localStorage.getItem(key);
    try {
      return JSON.parse(val);
    } catch (e) {
      return val;
    }
  }

  // Initialize on DOM ready
  function init() {
    initReactive();
    initEvents();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Public API
  window.SL = {
    setState,
    getState,
    watch,
    createElement,
    delegate,
    fetchData,
    saveLocal,
    loadLocal,
    renderTemplate
  };
})();
`
