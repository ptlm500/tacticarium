import { api } from './client';
import { MissionPack, Mission, Secondary, Gambit, MissionRule, ChallengerCard } from '../types/mission';

export const missionsApi = {
  listPacks: () => api.get<MissionPack[]>('/api/mission-packs'),
  listMissions: (packId: string) =>
    api.get<Mission[]>(`/api/mission-packs/${packId}/missions`),
  listSecondaries: (packId: string) =>
    api.get<Secondary[]>(`/api/mission-packs/${packId}/secondaries`),
  listGambits: (packId: string) =>
    api.get<Gambit[]>(`/api/mission-packs/${packId}/gambits`),
  listRules: (packId: string) =>
    api.get<MissionRule[]>(`/api/mission-packs/${packId}/rules`),
  listChallengerCards: (packId: string) =>
    api.get<ChallengerCard[]>(`/api/mission-packs/${packId}/challenger-cards`),
};
