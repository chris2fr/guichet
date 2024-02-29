/**
* This file was @generated using pocketbase-typegen
*/

import type PocketBase from 'pocketbase'
import type { RecordService } from 'pocketbase'

export enum Collections {
	MyStructures = "myStructures",
	NotYetMyStructures = "notYetMyStructures",
	Structures = "structures",
	Users = "users",
}

// Alias types for improved usability
export type IsoDateString = string
export type RecordIdString = string
export type HTMLString = string

// System fields
export type BaseSystemFields<T = never> = {
	id: RecordIdString
	created: IsoDateString
	updated: IsoDateString
	collectionId: string
	collectionName: Collections
	expand?: T
}

export type AuthSystemFields<T = never> = {
	email: string
	emailVisibility: boolean
	username: string
	verified: boolean
} & BaseSystemFields<T>

// Record types for each collection

export type MyStructuresRecord = {
	description?: HTMLString
	logo?: string
	name?: string
	users?: RecordIdString[]
}

export type NotYetMyStructuresRecord = {
	description?: HTMLString
	logo?: string
	name?: string
	users?: RecordIdString[]
}

export type StructuresRecord = {
	description?: HTMLString
	logo?: string
	name?: string
	users?: RecordIdString[]
}

export type UsersRecord = {
	avatar?: string
	canadmin?: boolean
	dn?: string
	name?: string
	structures?: RecordIdString[]
}

// Response types include system fields and match responses from the PocketBase API
export type MyStructuresResponse<Texpand = unknown> = Required<MyStructuresRecord> & BaseSystemFields<Texpand>
export type NotYetMyStructuresResponse<Texpand = unknown> = Required<NotYetMyStructuresRecord> & BaseSystemFields<Texpand>
export type StructuresResponse<Texpand = unknown> = Required<StructuresRecord> & BaseSystemFields<Texpand>
export type UsersResponse<Texpand = unknown> = Required<UsersRecord> & AuthSystemFields<Texpand>

// Types containing all Records and Responses, useful for creating typing helper functions

export type CollectionRecords = {
	myStructures: MyStructuresRecord
	notYetMyStructures: NotYetMyStructuresRecord
	structures: StructuresRecord
	users: UsersRecord
}

export type CollectionResponses = {
	myStructures: MyStructuresResponse
	notYetMyStructures: NotYetMyStructuresResponse
	structures: StructuresResponse
	users: UsersResponse
}

// Type for usage with type asserted PocketBase instance
// https://github.com/pocketbase/js-sdk#specify-typescript-definitions

export type TypedPocketBase = PocketBase & {
	collection(idOrName: 'myStructures'): RecordService<MyStructuresResponse>
	collection(idOrName: 'notYetMyStructures'): RecordService<NotYetMyStructuresResponse>
	collection(idOrName: 'structures'): RecordService<StructuresResponse>
	collection(idOrName: 'users'): RecordService<UsersResponse>
}
