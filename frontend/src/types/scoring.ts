export type PrimaryScoringSlot = "end_of_command_phase" | "end_of_battle_round" | "end_of_turn";

export const PRIMARY_SCORING_SLOT_LABELS: Record<PrimaryScoringSlot, string> = {
  end_of_command_phase: "End of Command Phase",
  end_of_battle_round: "End of Battle Round",
  end_of_turn: "End of Turn",
};
