import {getTranslations} from 'next-intl/server';

export default async function ServerComponent() {
  const t = await getTranslations('About');
  
  return (
    <div>
      <h1>{t('title')}</h1>
      <p>{t('description')}</p>
      {/* This key is not declared in the JSON file */}
      <p>{t('undeclaredKey')}</p>
    </div>
  );
} 