/**
 * Mobile Menu Module
 * Handles hamburger menu toggle for mobile/tablet viewports.
 */

export function initMobileMenu() {
  const burger = document.querySelector('[data-mobile-menu-toggle]');
  const menu = document.querySelector('[data-mobile-menu]');

  if (!burger || !menu) return;

  burger.addEventListener('click', () => {
    const isExpanded = burger.getAttribute('aria-expanded') === 'true';
    burger.setAttribute('aria-expanded', String(!isExpanded));
    menu.classList.toggle('is-open');

    // Update burger label
    const openLabel = burger.querySelector('[data-label-open]');
    const closeLabel = burger.querySelector('[data-label-close]');

    if (openLabel && closeLabel) {
      openLabel.hidden = !isExpanded;
      closeLabel.hidden = isExpanded;
    }
  });

  // Close menu on escape key
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && menu.classList.contains('is-open')) {
      burger.setAttribute('aria-expanded', 'false');
      menu.classList.remove('is-open');
      burger.focus();
    }
  });

  // Close menu when clicking outside
  document.addEventListener('click', (event) => {
    if (
      menu.classList.contains('is-open') &&
      !menu.contains(event.target) &&
      !burger.contains(event.target)
    ) {
      burger.setAttribute('aria-expanded', 'false');
      menu.classList.remove('is-open');
    }
  });
}
