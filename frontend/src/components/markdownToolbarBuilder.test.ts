import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import {
  buildToolbarFromSettings,
  type MinimalGroupBuilder,
  type MinimalItemGroup,
  type ToolbarItemConfig,
} from './markdownToolbarBuilder';
import type { MarkdownMenuSettings } from '../lib/stores/markdownMenuSettings';

// --- Fake builder ------------------------------------------------------------

interface AddItemCall {
  key: string;
  config: ToolbarItemConfig;
}

interface FakeBuilder extends MinimalGroupBuilder<ToolbarItemConfig> {
  groupCalls: string[];
  itemCalls: AddItemCall[];
}

function makeFakeBuilder(): FakeBuilder {
  const itemCalls: AddItemCall[] = [];
  const groupCalls: string[] = [];

  const group: MinimalItemGroup<ToolbarItemConfig> = {
    addItem(key, config) {
      itemCalls.push({ key, config });
      return group;
    },
  };

  return {
    groupCalls,
    itemCalls,
    addGroup(key: string) {
      groupCalls.push(key);
      return group;
    },
  };
}

function settings(overrides: Partial<MarkdownMenuSettings> = {}): MarkdownMenuSettings {
  return {
    bold: false,
    italic: false,
    strikethrough: false,
    code: false,
    link: false,
    latex: false,
    ...overrides,
  } as MarkdownMenuSettings;
}

// --- Tests -------------------------------------------------------------------

describe('buildToolbarFromSettings', () => {
  describe('AC-1: pure function contract', () => {
    it('returns distinct callback functions on successive invocations with the same input', () => {
      const s = settings({ bold: true, italic: true });
      const cb1 = buildToolbarFromSettings(s);
      const cb2 = buildToolbarFromSettings(s);

      expect(typeof cb1).toBe('function');
      expect(typeof cb2).toBe('function');
      expect(cb1).not.toBe(cb2);
    });

    it('does not mutate the settings object (frozen input is accepted without TypeError)', () => {
      const s = Object.freeze(
        settings({ bold: true, italic: true, strikethrough: true, code: true, link: true, latex: true }),
      );
      const snapshot = { ...s };

      const callback = buildToolbarFromSettings(s);
      const fake = makeFakeBuilder();

      expect(() => callback(fake)).not.toThrow();
      expect({ ...s }).toEqual(snapshot);
    });
  });

  describe('AC-2: all six enabled -> six addItem calls in fixed order', () => {
    it('calls addGroup("toolbar") exactly once and addItem for every key in canonical order', () => {
      const s = settings({
        bold: true,
        italic: true,
        strikethrough: true,
        code: true,
        link: true,
        latex: true,
      });
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(s)(fake);

      expect(fake.groupCalls).toEqual(['toolbar']);
      expect(fake.itemCalls.map(c => c.key)).toEqual([
        'bold',
        'italic',
        'strikethrough',
        'code',
        'link',
        'latex',
      ]);
    });
  });

  describe('AC-3: all disabled -> empty toolbar group', () => {
    it('still creates the toolbar group once and never calls addItem', () => {
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(settings())(fake);

      expect(fake.groupCalls).toEqual(['toolbar']);
      expect(fake.itemCalls).toHaveLength(0);
    });
  });

  describe('AC-4: partial selection -> only enabled keys, in order', () => {
    it('{bold, strikethrough, latex} enabled adds exactly those three in canonical order', () => {
      const s = settings({ bold: true, strikethrough: true, latex: true });
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(s)(fake);

      expect(fake.groupCalls).toEqual(['toolbar']);
      expect(fake.itemCalls.map(c => c.key)).toEqual(['bold', 'strikethrough', 'latex']);
    });

    it('{bold, latex} enabled adds exactly two items (BDD scenario: only bold and latex)', () => {
      const s = settings({ bold: true, latex: true });
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(s)(fake);

      expect(fake.itemCalls).toHaveLength(2);
      expect(fake.itemCalls.map(c => c.key)).toEqual(['bold', 'latex']);
    });
  });

  describe('AC-5: every enabled item carries icon, active, onRun', () => {
    it('each addItem config has truthy icon and function active/onRun', () => {
      const s = settings({
        bold: true,
        italic: true,
        strikethrough: true,
        code: true,
        link: true,
        latex: true,
      });
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(s)(fake);

      expect(fake.itemCalls).toHaveLength(6);
      for (const call of fake.itemCalls) {
        expect(call.config.icon).toBeTruthy();
        expect(typeof call.config.active).toBe('function');
        expect(typeof call.config.onRun).toBe('function');
      }
    });

    it('placeholder active/onRun do not throw when invoked', () => {
      const s = settings({ bold: true });
      const fake = makeFakeBuilder();

      buildToolbarFromSettings(s)(fake);

      const { active, onRun } = fake.itemCalls[0].config;
      expect(() => active({})).not.toThrow();
      expect(() => onRun({})).not.toThrow();
      expect(active({})).toBe(false);
    });
  });

  describe('AC-6: top-of-file strategy comment is declared', () => {
    it('first line declares primary or fallback strategy', () => {
      // vitest runs from the frontend/ package root; resolve relative to cwd.
      const sourcePath = resolve(process.cwd(), 'src/components/markdownToolbarBuilder.ts');
      const src = readFileSync(sourcePath, 'utf8');
      const firstLine = src.split('\n', 1)[0];
      expect(firstLine).toMatch(/^\/\/ Strategy: (primary|fallback)/);
    });
  });
});
