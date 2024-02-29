import { client } from "$lib/pocketbase";
import type { StructuresRecord, UsersRecord } from "$lib/pocketbase/generated-types";
import type { PageLoad } from "./$types";

export const load: PageLoad = async function ({ params: { slug: id } }) {
  const user: UsersRecord =
    id === "new"
      ? ({} as UsersRecord)
      : (await client.collection("users").getOne(id, {expand: 'structures'}));

    // const structures: StructuresRecord[] = await client.collection("structures").getFullList({
    //   filter: "structures_via_users.id != " + id,
    // })
  return {
    user,
  };
};
