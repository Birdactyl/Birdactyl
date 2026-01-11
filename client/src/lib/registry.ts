import { ComponentType, LazyExoticComponent, lazy } from 'react';

type Guard = 'auth' | 'admin' | null;

export interface PageRegistration {
  path: string;
  component: LazyExoticComponent<ComponentType<any>>;
  guard?: Guard;
}

export interface SidebarItem {
  id: string;
  label: string;
  icon: string;
  href: string;
  section: 'nav' | 'platform' | 'admin';
  order: number;
  guard?: Guard;
  children?: { label: string; href: string }[];
}

class Registry {
  private pages: PageRegistration[] = [];
  private sidebarItems: SidebarItem[] = [];

  registerPage(page: PageRegistration) {
    this.pages.push(page);
  }

  registerPages(pages: PageRegistration[]) {
    this.pages.push(...pages);
  }

  registerSidebarItem(item: SidebarItem) {
    this.sidebarItems.push(item);
  }

  registerSidebarItems(items: SidebarItem[]) {
    this.sidebarItems.push(...items);
  }

  getPages(guard?: Guard): PageRegistration[] {
    if (guard === undefined) return this.pages;
    return this.pages.filter(p => p.guard === guard);
  }

  getSidebarItems(section?: SidebarItem['section']): SidebarItem[] {
    const items = section 
      ? this.sidebarItems.filter(i => i.section === section)
      : this.sidebarItems;
    return items.sort((a, b) => a.order - b.order);
  }
}

export const registry = new Registry();

export const lazyPage = (importFn: () => Promise<{ default: ComponentType<any> }>) => lazy(importFn);
