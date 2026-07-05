/**
 * SCO terminal-configuration layout profile, per
 * docs/sco-terminal-implementation-design.md §2 ("Layout profiles (H/V/HD) as terminal
 * configuration"). M1 renders a single workflow with a profile class on the shell
 * (`data-layout="v"`) and minimal CSS differences; full visual polish per profile is not M1.
 */
export type LayoutProfile = 'h' | 'v' | 'hd';

const LAYOUT_PROFILES: readonly LayoutProfile[] = ['h', 'v', 'hd'];
const DEFAULT_LAYOUT_PROFILE: LayoutProfile = 'h';

export function parseLayoutProfile(value: string | undefined | null): LayoutProfile {
  if (value && (LAYOUT_PROFILES as readonly string[]).includes(value)) {
    return value as LayoutProfile;
  }
  return DEFAULT_LAYOUT_PROFILE;
}
