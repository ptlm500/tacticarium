import type { components } from "../../../shared/api.generated";

type Schemas = components["schemas"];

/** A faction with a guaranteed id (present when read from the API). */
export type Faction = Schemas["Faction"] & { id: string };

/** A detachment with a guaranteed id. */
export type Detachment = Schemas["Detachment"] & { id: string };

/** A stratagem with a guaranteed id. */
export type Stratagem = Schemas["Stratagem"] & { id: string };
