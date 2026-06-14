import { api } from "./client";
import {
  ForceDisposition,
  Mission,
  MissionMatchup,
  MissionCard,
  DeploymentPattern,
} from "../types/mission";

export const missionsApi = {
  listForceDispositions: () => api.get<ForceDisposition[]>("/api/force-dispositions"),
  listMissions: () => api.get<Mission[]>("/api/missions"),
  listMissionMatchups: () => api.get<MissionMatchup[]>("/api/mission-matchups"),
  listSecondaryCards: () => api.get<MissionCard[]>("/api/secondary-cards"),
  listDeploymentPatterns: () => api.get<DeploymentPattern[]>("/api/deployment-patterns"),
};
