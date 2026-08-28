export interface RoleSlot {
  role: string;
  label: string;
  isReply?: boolean;
  order: number;
}

export const BP_ROLES: Record<string, RoleSlot[]> = {
  OG: [
    { role: 'PM', label: 'Prime Minister (PM)', order: 1 },
    { role: 'DPM', label: 'Deputy Prime Minister (DPM)', order: 2 },
  ],
  OO: [
    { role: 'LO', label: 'Leader of Opposition (LO)', order: 1 },
    { role: 'DLO', label: 'Deputy Leader of Opposition (DLO)', order: 2 },
  ],
  CG: [
    { role: 'MG', label: 'Member for Government (MG)', order: 1 },
    { role: 'GW', label: 'Government Whip (GW)', order: 2 },
  ],
  CO: [
    { role: 'MO', label: 'Member for Opposition (MO)', order: 1 },
    { role: 'OW', label: 'Opposition Whip (OW)', order: 2 },
  ],
};

export const TWO_TEAM_ROLES: Record<string, RoleSlot[]> = {
  Gov: [
    { role: '1G', label: '1st Government (1G)', order: 1 },
    { role: '2G', label: '2nd Government (2G)', order: 2 },
    { role: '3G', label: '3rd Government (3G)', order: 3 },
    { role: 'GR', label: 'Government Reply (GR)', isReply: true, order: 4 },
  ],
  Opp: [
    { role: '1O', label: '1st Opposition (1O)', order: 1 },
    { role: '2O', label: '2nd Opposition (2O)', order: 2 },
    { role: '3O', label: '3rd Opposition (3O)', order: 3 },
    { role: 'OR', label: 'Opposition Reply (OR)', isReply: true, order: 4 },
  ],
};

/**
 * Returns the expected speech role slots for a given debate side.
 * Supports BP (OG, OO, CG, CO), 2-team formats (Gov, Opp, Prop, Aff, Neg), and fallback generic slots.
 */
export function getRoleSlotsForSide(side: string, totalTeamsInDebate: number = 4): RoleSlot[] {
  const norm = (side || '').trim().toUpperCase();
  if (BP_ROLES[norm]) {
    return BP_ROLES[norm];
  }
  if (
    norm === 'PROP' ||
    norm === 'AFF' ||
    norm === 'GOV' ||
    norm === 'GOVERNMENT' ||
    norm === 'PROPOSITION' ||
    norm === 'AFFIRMATIVE' ||
    norm === '1G' ||
    norm === 'G' ||
    norm === 'AFFIRMATIVE TEAM'
  ) {
    return TWO_TEAM_ROLES.Gov;
  }
  if (
    norm === 'OPP' ||
    norm === 'NEG' ||
    norm === 'OPPOSITION' ||
    norm === 'NEGATIVE' ||
    norm === '1O' ||
    norm === 'O' ||
    norm === 'NEGATIVE TEAM'
  ) {
    return TWO_TEAM_ROLES.Opp;
  }

  // 2-team fallback
  if (totalTeamsInDebate === 2) {
    if (norm.startsWith('G') || norm.startsWith('P') || norm.startsWith('A') || norm === '1') {
      return TWO_TEAM_ROLES.Gov;
    }
    return TWO_TEAM_ROLES.Opp;
  }

  // Default generic 2 speakers
  return [
    { role: `${side} 1`, label: `${side} 1st Speaker`, order: 1 },
    { role: `${side} 2`, label: `${side} 2nd Speaker`, order: 2 },
  ];
}
