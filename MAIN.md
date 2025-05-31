# Thesis

## Initial Thesis Topic Message
I sent:

> Lecturer: David Datuashvili
> Topic: "Unseen by eyes Machine Learning: A Multi-Layered Framework for E-Commerce Security, Personalization, and Monitoring"
> The topic extends the original idea of ​​network anomaly detection and adds analysis of several different approaches to web application deployment.

The topic can't be modified, same goes with the title.

Original message:

> ლექტორი: დავით დათუაშვილი
> თემა: "შეუმჩნეველი მანქანური სწავლება: მრავალფენიანი ჩარჩო ელექტრონული კომერციის უსაფრთხოებისთვის, პერსონალიზაციისთვის და მონიტორინგისთვის"
> თემა განავრცობს თავდაპირველ იდეას ქსელის ანომალიების აღმოჩენასთან დაკავშირებით და ამატებს რამდენიმე, განსხვავებული მიდგომის ანალიზს ვებ-აპლიკაციის წყობაში.

Initially, I thought about creating a Full-Stack E-Commerce Web Application that would display 3 cases of using Machine Learning in E-Commerce applications.
There were: Request Anomaly Detection + Prevention, Recommendations, Trend Analysis. General idea was to only use simplistic
classification and regression for all 3.

After a while, the general idea shifted to creating something more complex. I, forgetting that I can't really meddle with topic anymore, modified the project by removing some concepts and adding a lot of new ones.

This turned into a disaster quite quickly - Backend layer became too large, Frontend was way to cumbersome to set up while fitting everything, and so on.

Currently, I don't have a Frontend layer anymore. I fully removed it. I have a Database layer (PostgreSQL v15), a Backend
layer (Golang + Fiber) and a Machine Learning layer (Python + Scikit-Learn + Jupyter + FastAPI).

I also have a Thesis written that currently has 18 pages of content. The regulations state that the thesis should be between 20 and 30 pages.

Thesis itself is also monitored very strictly on subjects of plagiarism and AI usage. I don't really have any plagiarism going
on, because the paper itself only contains information about my project, but it's currently completely written by AI.

## Problems

### 1
I want to re-add one Machine Learning functionality to the system - `Request Analysis for Anomalies`. This will remove the issue of the thesis not corresponding to its title.

Including the newly added `Request Analysis for Anomalies`, there will be 7 Machine Learning applications used in the project:
1. `Request Analysis for Anomalies`
2. `Sentiment Analysis`
3. `Auto-Tagging`
4. `Smart Discounts`
5. `Product Recommendations`
6. `Enhanced Search`
7. `Trends Analysis`

Yes, it seems too much, but since I've removed the frontend and previous 6 applications work fine and only require a little tuning, I think this number of applications is fine.

### 2
Each Machine Learning model performs fine in current conditions, but there's one issue - They lack reasoning.

For example, there might be a question: "How does the enhanced search model influence the searching process and why is it useful?!"
To these kinds of questions, there's no answer. Each Machine Learning model works, but how, why, why not questions are not answered.

### 3
As I've already mentioned, thesis paper will be strictly checked for plagiarism and AI usage. Because of that, I want you to
only provide bullet points in large amounts instead of plain text for the thesis paper. I will then transform these bullet
points into sentences and text myself to get rid of AI usage claims.

The thesis' structure itself should also be modified. Currently, thesis only describes what the project does in words, but,
realistically, it should be a guidance and observation about the topic mentioned in the title. It's a research paper. Because
of that, I want you to also add markers about: Where to insert the image, Where to insert the tests, Where to provide personal
opinion, etc.

### 4
Since the removal of the Frontend layer, there's one issue - How to visualize my project. The issue here is not only in visualizing specific Machine Learning model traits or Backend API responses, because I can do that using Swagger - and in Swagger, I can freely provide structured responses to the judges, but it sounds cumbersome, will be hard to explain and lame.

Because of that, I want to create a presentation like feel to my project. I will still use Swagger for API documentation for
both Backend and Machine Learning layers, but I also need charts, graphs, plots, etc. I can use Jupyter notebook for that.

I also need order. By order, I mean I want my thesis paper and project to not feel out of sync. I don't want to move around
too much and follow a defined order in my presentation, as well as, in my thesis paper.

### 5
I need to store larger amounts of mock data.

Currently, I don't have many products, users, user interactions, etc. seeded into the database. Because of that, I can't
display the capabilities of my Machine Learning models fully. The validity of my research will also be under the question
if the data that the models are trained on will be too small.

### 6
Since most of this project was created by AI, I haven't implemented most of the stuff myself.

This is not an issue for the Backend or Database layers, because I am a Full-Stack Developer, but concerning Machine
Learning layer, this becomes an issue. Because Machine Learning layer is the main part of the thesis, this issue is even more
glaring.

In order to get out of this situation unharmed, I want to have a guidance about every single Machine Learning model that is
implemented in the project separately. The guidance should include:
1. Model description
2. Technologies used (SVM, Classification, Regression, PSA, Regularization, KNN, Supervised, Unsupervised, etc.)
3. Machine Learning techniques used (How data is organized, Data preparation, Avoiding overfitting or underfitting, Encodings, Model evaluation, Model tuning, etc.)
4. Why all the above mentioned (Why was the technology or technique chosen and used)
5. Custom implementation (Simple mock implementation)

## Issue #2: Lack of reasoning in Machine Learning models ✅ COMPLETED

**Problem**: The ML models work but lack reasoning and explainability. Users and evaluators can't understand how or why the models make specific decisions.

**Solution Implemented**: 
- ✅ **Comprehensive Explainability System**: Created a centralized explainability engine (`ml_service/models/explainability.py`) that provides detailed reasoning for all 7 ML models
- ✅ **Enhanced Model Reasoning**: Added detailed reasoning to all existing ML models:
  - **Recommendations**: Collaborative, content-based, and hybrid filtering with decision paths, confidence analysis, and algorithm breakdown
  - **Search**: TF-IDF semantic search with term matching analysis, similarity scoring, and query complexity assessment
  - **Anomaly Detection**: Multi-layered detection with method contribution analysis, risk assessment, and detection timeline
  - **Sentiment Analysis**: Lexicon-based analysis with word-level scoring and confidence metrics
  - **Auto-Tagging**: TF-IDF clustering with tag relevance scoring and extraction methodology
  - **Smart Discounts**: Performance-based optimization with ROI analysis and market factors
  - **Trends Analysis**: Linear regression with trend strength, seasonality detection, and forecast confidence

- ✅ **Explainability API**: New API endpoints (`/explainability/*`) providing:
  - Individual decision explanations with technical details and business impact
  - Batch explanation processing for multiple decisions
  - Model comparison and algorithm selection guidance
  - Confidence threshold analysis and business impact assessment

- ✅ **Six-Layer Explanation Framework**:
  1. **Model Overview**: Algorithm description, strengths, limitations, and use cases
  2. **Decision Reasoning**: Step-by-step decision path with contributing factors
  3. **Confidence Analysis**: Reliability indicators, uncertainty sources, and confidence scoring
  4. **Business Impact**: ROI potential, strategic value, and success metrics
  5. **Technical Details**: Implementation specifics, complexity analysis, and scalability notes
  6. **Improvement Suggestions**: Actionable recommendations for model enhancement

**Key Features**:
- **Transparent Decision Making**: Every ML prediction now includes detailed reasoning explaining how and why the decision was made
- **Confidence Scoring**: Multi-level confidence assessment (High/Medium/Low) with specific factors contributing to confidence levels
- **Business Context**: Each explanation includes business impact analysis and strategic value assessment
- **Technical Depth**: Comprehensive technical details including algorithm complexity, data requirements, and scalability considerations
- **Actionable Insights**: Specific suggestions for improving model performance and addressing limitations
- **Comparative Analysis**: Side-by-side algorithm comparison for informed decision making

**Academic Value**: This addresses the critical "black box" problem in ML systems, providing the transparency and interpretability required for academic evaluation and real-world deployment.

**Status**: ✅ **COMPLETED** - All 7 ML models now provide comprehensive reasoning and explainability