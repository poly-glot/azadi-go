/**
 * Azadi Financial Services - Main JS Entry Point
 * ES module that initializes all frontend modules.
 */

import { initMobileMenu } from './mobile-menu.js';
import { initInlineEdit } from './inline-edit.js';
import { initStripePayment } from './stripe-payment.js';

document.addEventListener('DOMContentLoaded', () => {
  initMobileMenu();
  initInlineEdit();
  initStripePayment();
});
