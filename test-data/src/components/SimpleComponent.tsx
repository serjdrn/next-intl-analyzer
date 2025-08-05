import {useTranslations} from 'next-intl';

function SimpleComponent() {
  const t = useTranslations('HomePage');
  
  return (
    <div>
      <h1>{t('title')}</h1>
      <p>{t('welcome')}</p>
      <button>{t('about')}</button>
    </div>
  );
}

export default SimpleComponent; 