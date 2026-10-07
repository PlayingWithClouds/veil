// Ergonomic, fully-typed data client over the Veil GraphQL API.
//
//   const veil = createVeilClient({ url: "http://localhost:8080/graphql" });
//   const scenes = await veil.scene.list({ limit: 50 });
//   const scene  = await veil.scene.get(id);
//
// Every method selects all scalar fields by default. Pass a genql selection as
// the last argument to pick fields or pull in relations:
//
//   await veil.scene.get(id, { id: true, title: true, tags: { name: true } });
//
// For anything outside plain CRUD (subscriptions, settings, custom mutations),
// use the raw genql client exposed as `veil.client`.

import {
  createClient,
  everything,
  type Client,
  type Query,
  type Mutation,
  type QueryGenqlSelection,
  type MutationGenqlSelection,
  type FieldsSelection,
} from "./generated/index";

export interface VeilClientOptions {
  /** GraphQL endpoint, e.g. http://localhost:8080/graphql */
  url: string;
  headers?: Record<string, string>;
  /** Custom fetch (defaults to global fetch). */
  fetch?: typeof fetch;
}

// --- generic plumbing ------------------------------------------------------

type QueryField = keyof Query & keyof QueryGenqlSelection;
type MutationField = keyof Mutation & keyof MutationGenqlSelection;

type Unwrap<T> = NonNullable<T> extends Array<infer E> ? E : NonNullable<T>;

// Arguments accepted by a field's genql selection, if any.
type ArgsOf<S> = NonNullable<S> extends { __args?: infer A } ? A : Record<string, never>;

const EVERYTHING = everything as { __scalar: true };

// --- resource factories ----------------------------------------------------

interface ReadConfig<One extends QueryField, Many extends QueryField> {
  one: One;
  many: Many;
}

interface WriteConfig<
  Create extends MutationField,
  Update extends MutationField,
  Remove extends MutationField,
> {
  create: Create;
  update: Update;
  remove: Remove;
}

function makeReaders<One extends QueryField, Many extends QueryField>(
  client: Client,
  cfg: ReadConfig<One, Many>,
) {
  type Model = Unwrap<Query[One]>;
  type ListModel = Unwrap<Query[Many]>;

  return {
    /** Fetch one record by id. */
    async get<S extends Record<string, any> = { __scalar: true }>(
      id: string,
      select?: S,
    ): Promise<FieldsSelection<Model, S> | null> {
      const request = { [cfg.one]: { __args: { id }, ...(select ?? EVERYTHING) } };
      const data = (await client.query(request as any)) as Record<string, any>;
      return data[cfg.one];
    },

    /** List/search records. */
    async list<S extends Record<string, any> = { __scalar: true }>(
      args?: ArgsOf<QueryGenqlSelection[Many]>,
      select?: S,
    ): Promise<FieldsSelection<ListModel, S>[]> {
      const request = { [cfg.many]: { __args: args ?? {}, ...(select ?? EVERYTHING) } };
      const data = (await client.query(request as any)) as Record<string, any>;
      return data[cfg.many];
    },
  };
}

function makeWriters<
  One extends QueryField,
  Create extends MutationField,
  Update extends MutationField,
  Remove extends MutationField,
>(client: Client, one: One, cfg: WriteConfig<Create, Update, Remove>) {
  type Model = Unwrap<Query[One]>;

  return {
    /** Create a record. */
    async create<S extends Record<string, any> = { __scalar: true }>(
      input: ArgsOf<MutationGenqlSelection[Create]> extends { input: infer I } ? I : never,
      select?: S,
    ): Promise<FieldsSelection<Model, S>> {
      const request = { [cfg.create]: { __args: { input }, ...(select ?? EVERYTHING) } };
      const data = (await client.mutation(request as any)) as Record<string, any>;
      return data[cfg.create];
    },

    /** Update a record by id. */
    async update<S extends Record<string, any> = { __scalar: true }>(
      id: string,
      input: ArgsOf<MutationGenqlSelection[Update]> extends { input: infer I } ? I : never,
      select?: S,
    ): Promise<FieldsSelection<Model, S>> {
      const request = { [cfg.update]: { __args: { id, input }, ...(select ?? EVERYTHING) } };
      const data = (await client.mutation(request as any)) as Record<string, any>;
      return data[cfg.update];
    },

    /** Delete a record by id. */
    async remove(id: string): Promise<boolean> {
      const request = { [cfg.remove]: { __args: { id } } };
      const data = (await client.mutation(request as any)) as Record<string, any>;
      return Boolean(data[cfg.remove]);
    },
  };
}

function crud<
  One extends QueryField,
  Many extends QueryField,
  Create extends MutationField,
  Update extends MutationField,
  Remove extends MutationField,
>(
  client: Client,
  read: ReadConfig<One, Many>,
  write: WriteConfig<Create, Update, Remove>,
) {
  return { ...makeReaders(client, read), ...makeWriters(client, read.one, write) };
}

// --- public client ---------------------------------------------------------

export type VeilClient = ReturnType<typeof createVeilClient>;

export function createVeilClient(options: VeilClientOptions) {
  const client = createClient({
    url: options.url,
    headers: options.headers as any,
    fetch: options.fetch,
  });

  return {
    /** Raw genql client for subscriptions, settings, and bespoke operations. */
    client,

    /** Stored playback/download sources for a scene. */
    stream: {
      /** List stored streams for a media record. */
      async forMedia<S extends Record<string, any> = { __scalar: true }>(
        mediaId: string,
        select?: S,
      ) {
        const request = { mediaStreams: { __args: { mediaId }, ...(select ?? EVERYTHING) } };
        const data = (await client.query(request as any)) as Record<string, any>;
        return data.mediaStreams as any[];
      },
      /** Add a stream source to a media record. */
      async create<S extends Record<string, any> = { __scalar: true }>(
        input: ArgsOf<MutationGenqlSelection["createStream"]> extends { input: infer I } ? I : never,
        select?: S,
      ) {
        const request = { createStream: { __args: { input }, ...(select ?? EVERYTHING) } };
        const data = (await client.mutation(request as any)) as Record<string, any>;
        return data.createStream;
      },
      /** Delete a stream by id. */
      async remove(id: string): Promise<boolean> {
        const request = { deleteStream: { __args: { id } } };
        const data = (await client.mutation(request as any)) as Record<string, any>;
        return Boolean(data.deleteStream);
      },
    },

    // Content entities are read-only over GraphQL — they are ingested via the
    // plugin pipeline, not created through the API.
    scene: makeReaders(client, { one: "scene", many: "scenes" }),
    performer: makeReaders(client, { one: "performer", many: "performers" }),
    studio: makeReaders(client, { one: "studio", many: "studios" }),
    tag: makeReaders(client, { one: "tag", many: "tags" }),
    gallery: makeReaders(client, { one: "gallery", many: "galleries" }),
    image: makeReaders(client, { one: "image", many: "images" }),
    collection: makeReaders(client, { one: "collection", many: "collections" }),
  };
}
