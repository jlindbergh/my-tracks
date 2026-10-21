const pluralRules = new Intl.PluralRules('en');

/** Picks the form that agrees with count, e.g. plural(2, 'activity was', 'activities were'). */
export function plural(count: number, one: string, other: string): string {
  return pluralRules.select(count) === 'one' ? one : other;
}

/** Formats a count with the form that agrees with it, e.g. formatCount(2, 'file', 'files') is "2 files". */
export function formatCount(count: number, one: string, other: string): string {
  return `${count} ${plural(count, one, other)}`;
}
