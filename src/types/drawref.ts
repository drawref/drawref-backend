export type TagEntry = {
  id: string;
  name: string;
  values: string[];
};

export type TagMap = Record<string, string[]>;

export type Category = {
  id: string;
  name: string;
  cover?: string;
  cover_id?: number;
  tags?: TagEntry[];
};

export type LocalImageInfo = {
  path: string;
};

export type Image = {
  id: number;
  path: string;
  author: string;
  author_url: string;
  tags?: TagMap;
};

export type SampleProviderEntry = {
  author: string;
  author_url: string;
  requirement: string;
  images: SampleCategoryEntry[];
};

export type SampleCategoryEntry = {
  category: string;
  images: SampleImage[];
};

export type SampleImage = {
  // path is relative to './public/sample'
  path: string;

  // db tags
  tags?: string[];
};
