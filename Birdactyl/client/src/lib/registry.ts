import { ComponentType, LazyExoticComponent, lazy } from 'react';

type Guard = 'auth' | 'admin' | null;

export interface PageRegistration {
  path: string;
  component: LazyExoticComponent<ComponentType<any>>;
  guard?: Guard;
}

/** Navigation item displayed in the top bar. */
export interface NavItem {
  id: string;
  label: string;
  icon: string;
  href: string;
  section: 'nav' | 'platform' | 'admin';
  order: number;
  guard?: Guard;
  children?: { label: string; href: string }[];
}

/** @deprecated Use NavItem instead. */
export type SidebarItem = NavItem;

class Registry {
  private pages: PageRegistration[] = [];
  private navItems: NavItem[] = [];

  registerPage(page: PageRegistration) {
    this.pages.push(page);
  }

  registerPages(pages: PageRegistration[]) {
    this.pages.push(...pages);
  }

  /** Register a top-bar navigation item. */
  registerNavItem(item: NavItem) {
    this.navItems.push(item);
  }

  /** Register multiple top-bar navigation items. */
  registerNavItems(items: NavItem[]) {
    this.navItems.push(...items);
  }

  /** @deprecated Use registerNavItem instead. */
  registerSidebarItem(item: NavItem) {
    this.registerNavItem(item);
  }

  /** @deprecated Use registerNavItems instead. */
  registerSidebarItems(items: NavItem[]) {
    this.registerNavItems(items);
  }

  getPages(guard?: Guard): PageRegistration[] {
    if (guard === undefined) return this.pages;
    return this.pages.filter(p => p.guard === guard);
  }

  /** Get top-bar navigation items, optionally filtered by section. */
  getNavItems(section?: NavItem['section']): NavItem[] {
    const items = section
      ? this.navItems.filter(i => i.section === section)
      : this.navItems;
    return items.sort((a, b) => a.order - b.order);
  }

  /** @deprecated Use getNavItems instead. */
  getSidebarItems(section?: NavItem['section']): NavItem[] {
    return this.getNavItems(section);
  }
}

export const registry = new Registry();

export const lazyPage = (importFn: () => Promise<{ default: ComponentType<any> }>) => lazy(importFn);
