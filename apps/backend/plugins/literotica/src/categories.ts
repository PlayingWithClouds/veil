interface Category {
  id: number;
  slug: string;
  name: string;
}

// The API names a category by number and page slug only; names are the ones the site's pages show.
const CATEGORIES: Category[] = [
  { id: 2, slug: "erotic-couplings", name: "Erotic Couplings" },
  { id: 4, slug: "exhibitionist-voyeur", name: "Exhibitionist & Voyeur" },
  { id: 5, slug: "fetish-stories", name: "Fetish" },
  { id: 6, slug: "gay-sex-stories", name: "Gay Male" },
  { id: 7, slug: "group-sex-stories", name: "Group Sex" },
  { id: 9, slug: "taboo-sex-stories", name: "Taboo/Incest" },
  { id: 10, slug: "interracial-erotic-stories", name: "Interracial Love" },
  { id: 11, slug: "lesbian-sex-stories", name: "Lesbian Sex" },
  { id: 12, slug: "loving-wives", name: "Loving Wives" },
  { id: 13, slug: "non-consent-stories", name: "Reluctance/NonConsent" },
  { id: 14, slug: "non-human-stories", name: "NonHuman" },
  { id: 15, slug: "adult-romance", name: "Romance" },
  { id: 16, slug: "masturbation-stories", name: "Toys & Masturbation" },
  { id: 26, slug: "mature-sex", name: "Mature" },
  { id: 27, slug: "celebrity-stories", name: "Fan Fiction & Celebrities" },
  { id: 29, slug: "mind-control", name: "Mind Control" },
  { id: 31, slug: "bdsm-stories", name: "BDSM" },
  { id: 33, slug: "erotic-novels", name: "Novels and Novellas" },
  { id: 37, slug: "anal-sex-stories", name: "Anal" },
  { id: 38, slug: "science-fiction-fantasy", name: "Sci-Fi & Fantasy" },
  { id: 40, slug: "first-time-sex-stories", name: "First Time" },
  { id: 48, slug: "transgender", name: "Transgender" },
  { id: 51, slug: "erotic-horror", name: "Erotic Horror" },
  { id: 58, slug: "crossdressing", name: "Crossdressing" },
];

/** The numeric id of a category page slug, or undefined when the slug isn't known. */
export function categoryIdOf(slug: string): number | undefined {
  const category = CATEGORIES.find((candidate) => candidate.slug === slug);
  if (category === undefined) {
    return undefined;
  }
  return category.id;
}

/** The display name of a category page slug; unknown slugs are title-cased. */
export function categoryName(slug: string): string {
  const category = CATEGORIES.find((candidate) => candidate.slug === slug);
  if (category !== undefined) {
    return category.name;
  }
  return slug
    .split("-")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
