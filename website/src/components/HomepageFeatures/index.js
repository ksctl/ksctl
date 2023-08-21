import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Lightweight',
    Svg: require('@site/static/img/lightweight.svg').default,
    description: (
      <>
        Ksctl is free from other CLI dependencies, it offers a lightweight solution for seamless user experience.
      </>
    ),
  },
  {
    title: 'Fast',
    Svg: require('@site/static/img/fast.svg').default,
    description: (
      <>
        Instantly generates clusters in a matter of seconds.
      </>
    ),
  },
  {
    title: 'Customize',
    Svg: require('@site/static/img/customize.svg').default,
    description: (
      <>
      Customize every cluster element, from node size, Kubernetes version, to included applications – all easily managed through our tool.     </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
