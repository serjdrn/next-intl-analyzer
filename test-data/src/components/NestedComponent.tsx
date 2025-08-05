import {useTranslations} from 'next-intl';

function NestedComponent() {
  const t = useTranslations('Common');
  
  return (
    <div>
      <button>{t('button.save')}</button>
      <button>{t('button.cancel')}</button>
      <nav>
        <a href="/">{t('navigation.home')}</a>
        <a href="/about">{t('navigation.about')}</a>
      </nav>
    </div>
  );
}

export default NestedComponent; 