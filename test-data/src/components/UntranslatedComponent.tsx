import { useTranslations } from 'next-intl';

function UntranslatedComponent() {
    const t = useTranslations('Common');

    return (
        <div>
            {/* These are properly translated */}
            <button>{t('button.save')}</button>
            <button>{t('button.cancel')}</button>

            {/* These are hardcoded untranslated strings */}
            <h1>Welcome to our application</h1>
            <p>This is a hardcoded string that should be translated</p>
            <button>Click here to continue</button>
            <span>Loading...</span>

            {/* More untranslated strings */}
            <div>
                <h2>About Us</h2>
                <p>This company was founded in 2020</p>
            </div>
        </div>
    );
}

export default UntranslatedComponent; 