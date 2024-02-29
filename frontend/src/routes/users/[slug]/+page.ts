import { client } from "$lib/pocketbase";
import type { PageLoad } from "./$types";

export const load: PageLoad = async function ({ url, params: { slug } }) {
  const { items } = await client
    .collection("users")
    .getList(undefined, undefined, {
      filter: `username="${slug}"`,
    });
  const [user] = items;

  return {
    user,
  };
};
